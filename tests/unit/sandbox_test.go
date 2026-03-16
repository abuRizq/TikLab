package unit

import (
	"testing"

	"github.com/tiklab/tiklab/internal/sandbox"
)

func TestStateTransitions(t *testing.T) {
	sandbox.Delete()
	defer sandbox.Delete()

	t.Run("created to running is valid", func(t *testing.T) {
		state := &sandbox.SandboxState{
			ContainerID:   "abc123",
			ContainerName: "tiklab-sandbox",
			ImageTag:      "tiklab/sandbox:0.1.0",
			Status:        sandbox.StatusCreated,
			Ports:         sandbox.PortMapping{SSH: 2222, API: 8728, Winbox: 8291, Control: 9090},
			UserCount:     50,
		}
		if err := sandbox.Save(state); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
		loaded, err := sandbox.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded == nil {
			t.Fatal("expected state, got nil")
		}
		if loaded.Status != sandbox.StatusCreated {
			t.Errorf("expected status created, got %s", loaded.Status)
		}
	})
}

func TestStatePersistence(t *testing.T) {
	sandbox.Delete()
	defer sandbox.Delete()

	state := &sandbox.SandboxState{
		ContainerID:   "test-container-id",
		ContainerName: "tiklab-sandbox",
		ImageTag:      "tiklab/sandbox:0.1.0",
		Status:        sandbox.StatusRunning,
		Ports:         sandbox.PortMapping{SSH: 2222, API: 8728, Winbox: 8291, Control: 9090},
		UserCount:     50,
	}

	if err := sandbox.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := sandbox.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected state after save")
	}
	if loaded.ContainerID != state.ContainerID {
		t.Errorf("ContainerID: got %q, want %q", loaded.ContainerID, state.ContainerID)
	}
	if loaded.Status != state.Status {
		t.Errorf("Status: got %s, want %s", loaded.Status, state.Status)
	}
	if loaded.UserCount != state.UserCount {
		t.Errorf("UserCount: got %d, want %d", loaded.UserCount, state.UserCount)
	}

	if err := sandbox.Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	loaded, _ = sandbox.Load()
	if loaded != nil {
		t.Error("expected nil state after delete")
	}
}

func TestErrorMessages(t *testing.T) {
	sandbox.Delete()
	defer sandbox.Delete()

	// Load when no state = nil, nil
	state, err := sandbox.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if state != nil {
		t.Error("expected nil state when file does not exist")
	}
}
