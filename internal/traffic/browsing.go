package traffic

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// RunBrowsing runs k6 HTTP GET requests in the namespace.
func RunBrowsing(ctx context.Context, nsExec func(name string, args ...string) *exec.Cmd, targetURL string) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	script := `import http from 'k6/http'; export const options = { vus: 1, duration: '1m', iterations: 100 }; export default function() { http.get('` + targetURL + `'); sleep(1); }`
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			cmd := nsExec("k6", "run", "--vus", "1", "--duration", "30s", "-")
			cmd.Stdin = strings.NewReader(script)
			_ = cmd.Run()
			time.Sleep(2 * time.Second)
		}
	}
}

// RunBrowsingCurl runs curl for HTTP GET when k6 is not available.
func RunBrowsingCurl(ctx context.Context, nsExec func(name string, args ...string) *exec.Cmd, targetURL string) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			cmd := nsExec("curl", "-s", "-o", "/dev/null", "-m", "5", targetURL)
			_ = cmd.Run()
		}
	}
}
