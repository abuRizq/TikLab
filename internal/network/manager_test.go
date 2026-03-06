package network

import (
	"runtime"
	"testing"
)

func TestManager_NonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("skipping non-Linux test on Linux")
	}
	dir := t.TempDir()
	mgr := NewManager(dir, ISPSmall())
	if err := mgr.Setup(); err == nil {
		t.Error("Setup should fail on non-Linux")
	}
}

func TestManager_ScaleTo_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires Linux")
	}
	dir := t.TempDir()
	mgr := NewManager(dir, ISPSmall())
	if err := mgr.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	defer mgr.Teardown()

	if err := mgr.ScaleTo(2); err != nil {
		t.Fatalf("ScaleTo 2: %v", err)
	}
	names, _ := mgr.ListNamespaces()
	if len(names) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(names))
	}

	if err := mgr.ScaleTo(1); err != nil {
		t.Fatalf("ScaleTo 1: %v", err)
	}
	names, _ = mgr.ListNamespaces()
	if len(names) != 1 {
		t.Errorf("expected 1 namespace, got %d", len(names))
	}

	if err := mgr.ScaleTo(0); err != nil {
		t.Fatalf("ScaleTo 0: %v", err)
	}
	names, _ = mgr.ListNamespaces()
	if len(names) != 0 {
		t.Errorf("expected 0 namespaces, got %d", len(names))
	}
}

func TestManager_PersistLoadState(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir, ISPSmall())
	if err := mgr.PersistState(42); err != nil {
		t.Fatalf("PersistState: %v", err)
	}
	n, err := mgr.LoadState()
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if n != 42 {
		t.Errorf("LoadState = %d, want 42", n)
	}
}

func TestManager_LoadState_NotExist(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir, ISPSmall())
	n, err := mgr.LoadState()
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if n != 0 {
		t.Errorf("LoadState = %d, want 0", n)
	}
}
