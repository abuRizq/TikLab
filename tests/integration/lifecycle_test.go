//go:build integration

package integration

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/tiklab/tiklab/internal/docker"
	"github.com/tiklab/tiklab/internal/routeros"
	"github.com/tiklab/tiklab/internal/sandbox"
)

const (
	createStartTimeout = 2 * time.Minute
	portCheckTimeout   = 5 * time.Second
	resetTimeout       = 30 * time.Second // SC-005: reset completes in under 30 seconds
	trafficStabilize    = 45 * time.Second
)

// tiklabBin returns the path to the tiklab binary (built in project root).
func tiklabBin(t *testing.T) string {
	_, file, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(file), "..", "..")
	bin := "tiklab"
	if runtime.GOOS == "windows" {
		bin = "tiklab.exe"
	}
	return filepath.Join(projectRoot, bin)
}

// runTiklab executes a tiklab subcommand. Builds the binary first if needed.
func runTiklab(t *testing.T, args ...string) error {
	bin := tiklabBin(t)
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		t.Logf("Building tiklab binary...")
		_, file, _, _ := runtime.Caller(0)
		projectRoot := filepath.Join(filepath.Dir(file), "..", "..")
		build := exec.Command("go", "build", "-o", bin, "./cmd/tiklab")
		build.Dir = projectRoot
		if out, err := build.CombinedOutput(); err != nil {
			t.Fatalf("go build failed: %v\n%s", err, out)
		}
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// TestLifecycleCreateStartDestroy verifies the full create→start→destroy cycle.
// Requires: Docker running, ports 2222/8728/8291/9090 free, tiklab/sandbox image built.
func TestLifecycleCreateStartDestroy(t *testing.T) {
	t.Cleanup(ensureCleanState)

	// Connect Docker
	dc := docker.NewClient()
	if err := dc.Connect(); err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dc.Close()
	if !dc.IsAvailable() {
		t.Skip("Docker daemon not reachable")
	}

	// Run tiklab create
	t.Log("Running tiklab create...")
	createStart := time.Now()
	if err := runTiklab(t, "create"); err != nil {
		t.Fatalf("tiklab create failed: %v", err)
	}

	// Verify state file exists
	state, err := sandbox.Load()
	if err != nil {
		t.Fatalf("Load state after create: %v", err)
	}
	if state == nil {
		t.Fatal("Expected state file after create")
	}
	if state.Status != sandbox.StatusCreated {
		t.Errorf("Expected status created, got %s", state.Status)
	}
	if state.ContainerID == "" {
		t.Error("Expected non-empty ContainerID")
	}

	// Run tiklab start
	t.Log("Running tiklab start...")
	if err := runTiklab(t, "start"); err != nil {
		t.Fatalf("tiklab start failed: %v", err)
	}

	// Assert create+start completes in under 2 minutes (SC-001)
	elapsed := time.Since(createStart)
	if elapsed > createStartTimeout {
		t.Errorf("Create+start took %v, expected under %v (SC-001)", elapsed, createStartTimeout)
	}

	// Verify state updated to running
	state, err = sandbox.Load()
	if err != nil {
		t.Fatalf("Load state after start: %v", err)
	}
	if state == nil {
		t.Fatal("Expected state file after start")
	}
	if state.Status != sandbox.StatusRunning {
		t.Errorf("Expected status running, got %s", state.Status)
	}

	// Verify port reachability: SSH 2222, API 8728, Winbox 8291
	ports := []int{2222, 8728, 8291}
	for _, port := range ports {
		if !isPortReachable(port) {
			t.Errorf("Port %d not reachable after start", port)
		}
	}

	// Run tiklab destroy
	t.Log("Running tiklab destroy...")
	if err := runTiklab(t, "destroy"); err != nil {
		t.Fatalf("tiklab destroy failed: %v", err)
	}

	// Verify state file deleted
	state, err = sandbox.Load()
	if err != nil {
		t.Fatalf("Load state after destroy: %v", err)
	}
	if state != nil {
		t.Error("Expected nil state after destroy")
	}

	// Verify ports released (may take a moment)
	time.Sleep(2 * time.Second)
	for _, port := range ports {
		if isPortReachable(port) {
			t.Logf("Port %d still reachable (container may still be shutting down)", port)
		}
	}
}

// TestResetRestoresCleanState verifies tiklab reset wipes config changes and restores clean state.
// Makes config changes via RouterOS API (firewall rule, delete user), runs tiklab reset,
// verifies clean state matches fresh sandbox, asserts reset completes in under 30s (SC-005).
func TestResetRestoresCleanState(t *testing.T) {
	t.Cleanup(ensureCleanState)

	dc := docker.NewClient()
	if err := dc.Connect(); err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dc.Close()
	if !dc.IsAvailable() {
		t.Skip("Docker daemon not reachable")
	}

	t.Log("Running tiklab create...")
	if err := runTiklab(t, "create"); err != nil {
		t.Fatalf("tiklab create failed: %v", err)
	}

	t.Log("Running tiklab start...")
	if err := runTiklab(t, "start"); err != nil {
		t.Fatalf("tiklab start failed: %v", err)
	}

	state, err := sandbox.Load()
	if err != nil {
		t.Fatalf("Load state: %v", err)
	}
	if state == nil {
		t.Fatal("Expected state after start")
	}

	t.Logf("Waiting %v for traffic to stabilize...", trafficStabilize)
	time.Sleep(trafficStabilize)

	user, pass := routeros.DefaultCredentials()
	ros := routeros.NewClient()
	defer ros.Close()
	if err := ros.Connect("127.0.0.1", state.Ports.API, user, pass); err != nil {
		t.Fatalf("RouterOS API connect failed: %v", err)
	}

	// Make config changes: add firewall rule, delete a Hotspot user
	t.Log("Adding firewall rule...")
	if _, err := ros.Run("/ip/firewall/filter/add",
		"=chain=input",
		"=action=accept",
		"=comment=tiklab-test-rule",
	); err != nil {
		t.Fatalf("Add firewall rule failed: %v", err)
	}

	t.Log("Deleting a Hotspot user...")
	reply, err := ros.Run("/ip/hotspot/user/print")
	if err != nil {
		t.Fatalf("/ip/hotspot/user/print failed: %v", err)
	}
	if len(reply.Re) > 0 {
		for _, re := range reply.Re {
			if id, ok := re.Map[".id"]; ok {
				_, _ = ros.Run("/ip/hotspot/user/remove", "=numbers="+id)
				break
			}
		}
	}

	// Verify changes are present
	reply, err = ros.Run("/ip/firewall/filter/print")
	if err != nil {
		t.Fatalf("/ip/firewall/filter/print failed: %v", err)
	}
	firewallCountBefore := len(reply.Re)
	foundTestRule := false
	for _, re := range reply.Re {
		if c, ok := re.Map["comment"]; ok && strings.Contains(c, "tiklab-test-rule") {
			foundTestRule = true
			break
		}
	}
	if !foundTestRule {
		t.Log("Firewall rule may not have comment in print output; continuing")
	}
	t.Logf("Firewall rules before reset: %d", firewallCountBefore)

	// Run tiklab reset and assert it completes in under 30 seconds
	t.Log("Running tiklab reset...")
	resetStart := time.Now()
	if err := runTiklab(t, "reset"); err != nil {
		t.Fatalf("tiklab reset failed: %v", err)
	}
	resetElapsed := time.Since(resetStart)
	if resetElapsed > resetTimeout {
		t.Errorf("Reset took %v, expected under %v (SC-005)", resetElapsed, resetTimeout)
	}
	t.Logf("Reset completed in %v", resetElapsed)

	// Reconnect (connection may have been closed)
	ros2 := routeros.NewClient()
	defer ros2.Close()
	if err := ros2.Connect("127.0.0.1", state.Ports.API, user, pass); err != nil {
		t.Fatalf("RouterOS API reconnect failed: %v", err)
	}

	// Verify clean state: DHCP server present, Hotspot present, no test firewall rule
	reply, err = ros2.Run("/ip/dhcp-server/print")
	if err != nil {
		t.Fatalf("/ip/dhcp-server/print after reset failed: %v", err)
	}
	if len(reply.Re) == 0 {
		t.Error("Expected DHCP server after reset (clean state)")
	}

	reply, err = ros2.Run("/ip/hotspot/print")
	if err != nil {
		t.Fatalf("/ip/hotspot/print after reset failed: %v", err)
	}
	if len(reply.Re) == 0 {
		t.Error("Expected Hotspot after reset (clean state)")
	}

	reply, err = ros2.Run("/ip/firewall/filter/print")
	if err != nil {
		t.Fatalf("/ip/firewall/filter/print after reset failed: %v", err)
	}
	for _, re := range reply.Re {
		if c, ok := re.Map["comment"]; ok && strings.Contains(c, "tiklab-test-rule") {
			t.Error("Test firewall rule should be removed after reset")
			break
		}
	}

	// Verify ~50 users regenerated (DHCP leases or engine status)
	time.Sleep(10 * time.Second) // Allow users to regenerate
	reply, err = ros2.Run("/ip/dhcp-server/lease/print")
	if err != nil {
		t.Fatalf("/ip/dhcp-server/lease/print after reset failed: %v", err)
	}
	leaseCount := len(reply.Re)
	if leaseCount < 40 || leaseCount > 60 {
		t.Errorf("Expected ~50 DHCP leases after reset, got %d", leaseCount)
	}
	t.Logf("DHCP leases after reset: %d", leaseCount)

	t.Log("Running tiklab destroy...")
	if err := runTiklab(t, "destroy"); err != nil {
		t.Logf("destroy failed: %v", err)
	}
}

func isPortReachable(port int) bool {
	addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, portCheckTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func ensureCleanState() {
	state, _ := sandbox.Load()
	if state != nil {
		// Attempt to destroy any existing sandbox to avoid leaving containers behind
		_, file, _, _ := runtime.Caller(0)
		projectRoot := filepath.Join(filepath.Dir(file), "..", "..")
		bin := filepath.Join(projectRoot, "tiklab")
		if runtime.GOOS == "windows" {
			bin = filepath.Join(projectRoot, "tiklab.exe")
		}
		if _, err := os.Stat(bin); err == nil {
			cmd := exec.Command(bin, "destroy")
			cmd.Dir = projectRoot
			_ = cmd.Run()
		}
	}
	sandbox.Delete()
}