package sandbox

import (
	"context"
	"fmt"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tiklab/tiklab/internal/debug"
	"github.com/tiklab/tiklab/internal/docker"
)

const (
	kvmBootTimeout       = 180 * time.Second
	emulationBootTimeout = 900 * time.Second
	kvmDetectDelay       = 5 * time.Second
)

// Manager orchestrates sandbox lifecycle operations.
type Manager struct {
	docker *docker.Client
}

// NewManager creates a new sandbox manager.
func NewManager(d *docker.Client) *Manager {
	return &Manager{docker: d}
}

// Create provisions a new sandbox environment.
func (m *Manager) Create(ctx context.Context, imageTag, containerName string) error {
	state, err := Load()
	if err != nil {
		return err
	}
	if state != nil {
		return fmt.Errorf("Sandbox already exists. Run `tiklab destroy` first")
	}

	ports := docker.DefaultPortMapping()
	containerID, err := m.docker.CreateContainer(ctx, containerName, imageTag, ports)
	if err != nil {
		return err
	}

	now := time.Now()
	state = &SandboxState{
		ContainerID:   containerID,
		ContainerName: containerName,
		ImageTag:      imageTag,
		Status:        StatusCreated,
		CreatedAt:     now,
		Ports: PortMapping{
			SSH:     ports.SSH,
			API:     ports.API,
			Winbox:  ports.Winbox,
			Control: ports.Control,
		},
		UserCount: 50,
	}
	if err := Save(state); err != nil {
		// Clean up container on state save failure to avoid orphan
		_ = m.docker.RemoveContainer(ctx, containerID, true)
		return err
	}
	return nil
}

// WaitForReadyFunc is called after the container starts to wait for RouterOS to boot.
// Receives host, API port, and boot timeout. Returns when ready or on timeout.
type WaitForReadyFunc func(ctx context.Context, host string, port int, timeout time.Duration) error

// Start activates a created sandbox.
// If waitForReady is non-nil, it is called after starting the container to wait for RouterOS.
func (m *Manager) Start(ctx context.Context, waitForReady WaitForReadyFunc) error {
	state, err := Load()
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("No sandbox found. Run `tiklab create` first")
	}
	if state.Status == StatusRunning {
		return fmt.Errorf("Sandbox is already running")
	}

	// #region agent log
	portAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(state.Ports.API))
	ln, portErr := net.Listen("tcp", portAddr)
	if portErr != nil {
		debug.Log("manager.go:Start", "port_in_use_BEFORE_start", map[string]interface{}{
			"port": state.Ports.API, "err": portErr.Error(),
		}, "H4")
	} else {
		ln.Close()
		debug.Log("manager.go:Start", "port_free_before_start", map[string]interface{}{
			"port": state.Ports.API,
		}, "H4")
	}
	// #endregion

	if err := m.docker.StartContainer(ctx, state.ContainerID); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// #region agent log
	status, exitCode, _ := m.docker.ContainerInspect(ctx, state.ContainerID)
	cid := state.ContainerID
	if len(cid) > 12 {
		cid = cid[:12]
	}
	debug.Log("manager.go:Start", "after_start_container", map[string]interface{}{
		"containerID": cid,
		"status":      status,
		"exitCode":    exitCode,
		"apiPort":     state.Ports.API,
	}, "H1")
	if status != "running" {
		logs, _ := m.docker.ContainerLogs(ctx, state.ContainerID)
		debug.Log("manager.go:Start", "container_not_running_logs", map[string]interface{}{
			"logs": tailStr(logs, 2000),
		}, "H1")
	}
	// #endregion

	bootTimeout := kvmBootTimeout
	emulationMode := false
	time.Sleep(kvmDetectDelay)
	earlyLogs, _ := m.docker.ContainerLogs(ctx, state.ContainerID)
	if strings.Contains(earlyLogs, "KVM not available") {
		emulationMode = true
		bootTimeout = emulationBootTimeout
		log.Printf("[sandbox] KVM not available — QEMU running in emulation mode.")
		log.Printf("[sandbox] Boot will be SLOW (timeout %ds). To fix, enable nested virtualization:", int(bootTimeout.Seconds()))
		if runtime.GOOS == "windows" {
			log.Printf("[sandbox]   1. Open PowerShell as Administrator")
			log.Printf("[sandbox]   2. Run: echo '[wsl2]\nnestedVirtualization=true' > $env:USERPROFILE\\.wslconfig")
			log.Printf("[sandbox]   3. Run: wsl --shutdown")
			log.Printf("[sandbox]   4. Restart Docker Desktop")
		} else {
			log.Printf("[sandbox]   Ensure KVM is available: sudo modprobe kvm && ls /dev/kvm")
		}
	}
	// #region agent log
	debug.Log("manager.go:Start", "kvm_detection", map[string]interface{}{
		"emulation":  emulationMode,
		"timeoutSec": int(bootTimeout.Seconds()),
	}, "H2")
	// #endregion

	if waitForReady != nil {
		// #region agent log
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			check := 0
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					check++
					st, ec, _ := m.docker.ContainerInspect(ctx, state.ContainerID)
					logs, _ := m.docker.ContainerLogs(ctx, state.ContainerID)
					debug.Log("manager.go:Start", "periodic_container_check", map[string]interface{}{
						"check": check, "status": st, "exitCode": ec,
						"logsTail": tailStr(logs, 2000),
					}, "H1")
				}
			}
		}()
		// #endregion

		if err := waitForReady(ctx, "127.0.0.1", state.Ports.API, bootTimeout); err != nil {
			// #region agent log
			close(done)
			status2, exitCode2, _ := m.docker.ContainerInspect(ctx, state.ContainerID)
			logs, _ := m.docker.ContainerLogs(ctx, state.ContainerID)
			debug.Log("manager.go:Start", "waitForReady_FAILED", map[string]interface{}{
				"status": status2, "exitCode": exitCode2, "err": err.Error(),
				"logsTail": tailStr(logs, 3000),
			}, "H1")
			// #endregion
			if emulationMode {
				return fmt.Errorf("RouterOS failed to boot: QEMU is running without KVM hardware acceleration.\n\n"+
					"  This makes boot extremely slow or impossible.\n\n"+
					"  Fix: Enable nested virtualization in WSL2:\n"+
					"    1. Open PowerShell as Administrator\n"+
					"    2. Run: echo '[wsl2]\\nnestedVirtualization=true' > $env:USERPROFILE\\.wslconfig\n"+
					"    3. Run: wsl --shutdown\n"+
					"    4. Restart Docker Desktop\n"+
					"    5. Run: tiklab destroy && tiklab create && tiklab start")
			}
			return err
		}
		// #region agent log
		close(done)
		// #endregion
	}

	now := time.Now()
	state.Status = StatusRunning
	state.StartedAt = &now
	return Save(state)
}

// Scale adjusts the number of simulated users.
func (m *Manager) Scale(ctx context.Context, n int) error {
	state, err := Load()
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("No sandbox found. Run `tiklab create` first")
	}
	if state.Status != StatusRunning {
		return fmt.Errorf("Sandbox is not running. Run `tiklab start` first")
	}
	if n < 1 {
		return fmt.Errorf("Minimum user count is 1")
	}
	if n > 500 {
		return fmt.Errorf("Maximum user count is 500")
	}

	state.UserCount = n
	return Save(state)
}

// Reset restores the sandbox to its original state.
func (m *Manager) Reset(ctx context.Context) error {
	state, err := Load()
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("No sandbox found. Run `tiklab create` first")
	}
	if state.Status != StatusRunning {
		return fmt.Errorf("Sandbox is not running. Run `tiklab start` first")
	}

	// Reset logic (config wipe, engine restart) is implemented in Phase 5.
	// For Phase 2, we only validate state.
	state.UserCount = 50
	return Save(state)
}

// Destroy removes all sandbox artifacts.
func (m *Manager) Destroy(ctx context.Context) error {
	state, err := Load()
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("No sandbox found. Nothing to destroy")
	}

	// Stop container if running
	_ = m.docker.StopContainer(ctx, state.ContainerID)

	// Remove container with volumes
	if err := m.docker.RemoveContainer(ctx, state.ContainerID, true); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return Delete()
}

// #region agent log
func tailStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return "..." + s[len(s)-maxLen:]
}

// #endregion
