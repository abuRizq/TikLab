package sandbox

import (
	"context"
	"fmt"
	"time"

	"github.com/tiklab/tiklab/internal/docker"
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
	return Save(state)
}

// WaitForReadyFunc is called after the container starts to wait for RouterOS to boot.
// Receives host and API port. Returns when ready or on timeout.
type WaitForReadyFunc func(ctx context.Context, host string, port int) error

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

	if err := m.docker.StartContainer(ctx, state.ContainerID); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	if waitForReady != nil {
		if err := waitForReady(ctx, "127.0.0.1", state.Ports.API); err != nil {
			return err
		}
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
