package traffic

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// RunHeavy runs iperf3 client in the namespace for sustained bandwidth.
func RunHeavy(ctx context.Context, nsExec func(name string, args ...string) *exec.Cmd, serverHost string, serverPort int) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	portStr := "5201"
	if serverPort > 0 {
		portStr = fmt.Sprintf("%d", serverPort)
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			cmd := nsExec("iperf3", "-c", serverHost, "-p", portStr, "-t", "20", "-P", "1")
			_ = cmd.Run()
		}
	}
}
