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
		return fmt.Errorf("sandbox already exists. Run `tiklab destroy` first")
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

// Start activates a created sandbox.
func (m *Manager) Start(ctx context.Context) error {
	state, err := Load()
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("no sandbox found. Run `tiklab create` first")
	}
	if state.Status == StatusRunning {
		return fmt.Errorf("sandbox is already running")
	}

	if err := m.docker.StartContainer(ctx, state.ContainerID); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
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
		return fmt.Errorf("no sandbox found. Run `tiklab create` first")
	}
	if state.Status != StatusRunning {
		return fmt.Errorf("sandbox is not running. Run `tiklab start` first")
	}
	if n < 1 {
		return fmt.Errorf("minimum user count is 1")
	}
	if n > 500 {
		return fmt.Errorf("maximum user count is 500")
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
		return fmt.Errorf("no sandbox found. Run `tiklab create` first")
	}
	if state.Status != StatusRunning {
		return fmt.Errorf("sandbox is not running. Run `tiklab start` first")
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
		return fmt.Errorf("no sandbox found. Nothing to destroy")
	}

	// Stop container if running
	_ = m.docker.StopContainer(ctx, state.ContainerID)

	// Remove container with volumes
	if err := m.docker.RemoveContainer(ctx, state.ContainerID, true); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return Delete()
}
