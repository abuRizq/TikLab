package network

import (
	"runtime"
	"testing"
)

func TestCreateBridge_NonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("skipping non-Linux test on Linux")
	}
	err := CreateBridge("br-test")
	if err == nil {
		t.Error("expected error on non-Linux")
	}
}

func TestBridgeExists_NonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("skipping non-Linux test on Linux")
	}
	if BridgeExists("br-test") {
		t.Error("bridge should not exist on non-Linux")
	}
}

func TestCreateDeleteBridge_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires Linux")
	}
	name := "br-sandbox-test"
	defer DeleteBridge(name)

	if err := CreateBridge(name); err != nil {
		t.Fatalf("CreateBridge: %v", err)
	}
	if !BridgeExists(name) {
		t.Error("bridge should exist")
	}
	if err := DeleteBridge(name); err != nil {
		t.Fatalf("DeleteBridge: %v", err)
	}
	if BridgeExists(name) {
		t.Error("bridge should be gone")
	}
}
