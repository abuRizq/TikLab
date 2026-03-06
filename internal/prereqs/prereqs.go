package prereqs

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Result holds the outcome of prerequisite checks.
type Result struct {
	OK      bool
	Docker  bool
	QEMU    bool
	Linux   bool
	Missing []string
}

// CheckAll runs all prerequisite checks and returns the result.
func CheckAll() (*Result, error) {
	r := &Result{OK: true}

	// Docker is required on all platforms
	if path, err := exec.LookPath("docker"); err == nil && path != "" {
		r.Docker = true
	} else {
		r.Missing = append(r.Missing, "docker")
		r.OK = false
	}

	// QEMU is required for CHR VM (Linux typically has it)
	if path, err := exec.LookPath("qemu-system-x86_64"); err == nil && path != "" {
		r.QEMU = true
	} else {
		r.Missing = append(r.Missing, "qemu-system-x86_64")
		r.OK = false
	}

	// Full sandbox (network namespaces, bridge) requires Linux
	r.Linux = runtime.GOOS == "linux"
	if !r.Linux {
		r.Missing = append(r.Missing, "linux (network namespaces)")
		r.OK = false
	}

	return r, nil
}

// Summary returns a human-readable summary of missing prerequisites.
func (r *Result) Summary() string {
	if r.OK {
		return "all checks passed"
	}
	return strings.Join(r.Missing, ", ")
}

// CheckDocker returns an error if Docker is not available.
func CheckDocker() error {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return errors.New("Docker is not running or not installed")
	}
	return nil
}

// RequireLinux returns an error if not running on Linux.
func RequireLinux() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("sandbox requires Linux (current: %s)", runtime.GOOS)
	}
	return nil
}
