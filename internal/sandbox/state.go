package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Status represents the sandbox lifecycle state.
type Status string

const (
	StatusCreated Status = "created"
	StatusRunning Status = "running"
)

// PortMapping defines host-to-container port mappings.
type PortMapping struct {
	SSH     int `json:"ssh"`
	API     int `json:"api"`
	Winbox  int `json:"winbox"`
	Control int `json:"control"`
}

// SandboxState represents the persisted state of the active sandbox instance.
type SandboxState struct {
	ContainerID   string      `json:"containerId"`
	ContainerName string      `json:"containerName"`
	ImageTag      string      `json:"imageTag"`
	Status        Status      `json:"status"`
	CreatedAt     time.Time   `json:"createdAt"`
	StartedAt     *time.Time  `json:"startedAt,omitempty"`
	Ports         PortMapping `json:"ports"`
	UserCount     int         `json:"userCount"`
}

// DefaultStatePath returns the default path for the state file (~/.tiklab/state.json).
func DefaultStatePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".tiklab", "state.json"), nil
}

// Load reads the sandbox state from the default state file.
// Returns nil, nil if the file does not exist (no sandbox).
func Load() (*SandboxState, error) {
	path, err := DefaultStatePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}
	var state SandboxState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}
	return &state, nil
}

// Save writes the sandbox state to the default state file.
// Creates the ~/.tiklab directory if it does not exist.
func Save(state *SandboxState) error {
	path, err := DefaultStatePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	return nil
}

// Delete removes the state file.
// Returns nil if the file does not exist.
func Delete() error {
	path, err := DefaultStatePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete state file: %w", err)
	}
	return nil
}
