package network

import (
	"runtime"
	"testing"
)

func TestCreateDeleteNamespace_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires Linux")
	}
	// Create bridge first
	if err := CreateBridge("br-ns-test"); err != nil {
		t.Fatalf("CreateBridge: %v", err)
	}
	defer DeleteBridge("br-ns-test")

	ns, err := CreateNamespace("br-ns-test", "sandbox-0001", 0)
	if err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	if ns.Name != "sandbox-0001" {
		t.Errorf("Name = %q", ns.Name)
	}

	names, _ := ListNamespaces()
	found := false
	for _, n := range names {
		if n == ns.Name {
			found = true
			break
		}
	}
	if !found {
		t.Error("namespace should be in list")
	}

	if err := DeleteNamespace(ns); err != nil {
		t.Fatalf("DeleteNamespace: %v", err)
	}
}
