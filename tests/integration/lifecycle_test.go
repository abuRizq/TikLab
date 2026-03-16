//go:build integration

package integration

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/tiklab/tiklab/internal/docker"
	"github.com/tiklab/tiklab/internal/sandbox"
)

const (
	createStartTimeout = 2 * time.Minute
	portCheckTimeout   = 5 * time.Second
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