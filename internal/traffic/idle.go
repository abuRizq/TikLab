package traffic

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// RunIdle runs ping and DNS in the namespace to keep the session active.
func RunIdle(ctx context.Context, nsExec func(name string, args ...string) *exec.Cmd, gateway string) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Ping gateway
			cmd := nsExec("ping", "-c", "1", "-W", "2", gateway)
			_ = cmd.Run()
			// DNS lookup (nslookup or getent)
			cmd = nsExec("getent", "hosts", "google.com")
			if cmd.Run() != nil {
				cmd = nsExec("nslookup", "google.com", gateway)
				_ = cmd.Run()
			}
		}
	}
}

// RunIdleOnce runs a single ping cycle (for testing).
func RunIdleOnce(nsExec func(name string, args ...string) *exec.Cmd, gateway string) error {
	cmd := nsExec("ping", "-c", "1", "-W", "2", gateway)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ping: %w: %s", err, string(out))
	}
	return nil
}
