package traffic

import (
	"context"
	"runtime"
	"testing"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine("192.168.88.1", []string{"sandbox-0000", "sandbox-0001"})
	if e.Gateway != "192.168.88.1" {
		t.Errorf("Gateway = %q", e.Gateway)
	}
	if len(e.Namespaces) != 2 {
		t.Errorf("Namespaces = %d", len(e.Namespaces))
	}
}

func TestEngine_Start_NonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("skipping non-Linux test on Linux")
	}
	e := NewEngine("192.168.88.1", nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := e.Start(ctx)
	if err == nil {
		t.Error("expected error on non-Linux")
	}
}
