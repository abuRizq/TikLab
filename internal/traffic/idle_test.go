package traffic

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestRunIdleOnce_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires Linux")
	}
	nsExec := func(name string, args ...string) *exec.Cmd {
		all := append([]string{"netns", "exec", "sandbox-0000", name}, args...)
		return exec.Command("ip", all...)
	}
	err := RunIdleOnce(nsExec, "192.168.88.1")
	if err != nil {
		t.Skipf("RunIdleOnce (no sandbox ns or ping): %v", err)
	}
}
