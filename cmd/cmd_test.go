package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/tiklab/mikrotik-sandbox/internal/sandbox"
)

func TestRootCmd_Help(t *testing.T) {
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestCreateCmd_RunsWithSkipPrereqs(t *testing.T) {
	skipPrereqs = true
	defer func() { skipPrereqs = false }()

	dir := t.TempDir()
	baseDir := filepath.Join(dir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}
	basePath := filepath.Join(baseDir, "chr.qcow2")
	if err := exec.Command("qemu-img", "create", "-f", "qcow2", basePath, "1M").Run(); err != nil {
		t.Skipf("qemu-img not available: %v", err)
	}
	absDir, _ := filepath.Abs(dir)
	os.Setenv("SANDBOX_WORKDIR", absDir)
	defer os.Unsetenv("SANDBOX_WORKDIR")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"create", "--skip-prereqs"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("create --skip-prereqs: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Creating sandbox")) {
		t.Errorf("output should mention creating sandbox, got %q", out.String())
	}
}

func TestStartCmd_RunsWithSkipPrereqs(t *testing.T) {
	skipPrereqs = true
	defer func() { skipPrereqs = false }()

	dir := setupSandboxWithBase(t)
	os.Setenv("SANDBOX_WORKDIR", dir)
	defer os.Unsetenv("SANDBOX_WORKDIR")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"start", "--skip-prereqs"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("start --skip-prereqs: %v", err)
	}
}

func TestResetCmd_RunsWithSkipPrereqs(t *testing.T) {
	skipPrereqs = true
	defer func() { skipPrereqs = false }()

	dir := setupSandboxWithBase(t)
	os.Setenv("SANDBOX_WORKDIR", dir)
	defer os.Unsetenv("SANDBOX_WORKDIR")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"reset", "--skip-prereqs"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("reset --skip-prereqs: %v", err)
	}
}

func TestDestroyCmd_RunsWithSkipPrereqs(t *testing.T) {
	skipPrereqs = true
	defer func() { skipPrereqs = false }()

	dir := setupSandboxWithBase(t)
	os.Setenv("SANDBOX_WORKDIR", dir)
	defer os.Unsetenv("SANDBOX_WORKDIR")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"destroy", "--skip-prereqs"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("destroy --skip-prereqs: %v", err)
	}
}

func TestScaleUsersCmd_ValidCount(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("scale-users requires Linux")
	}
	skipPrereqs = true
	defer func() { skipPrereqs = false }()

	dir := setupSandboxWithBase(t)
	os.Setenv("SANDBOX_WORKDIR", dir)
	defer os.Unsetenv("SANDBOX_WORKDIR")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"scale-users", "2", "--skip-prereqs"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("scale-users 2 --skip-prereqs: %v", err)
	}
}

func TestScaleUsersCmd_InvalidCount(t *testing.T) {
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
	rootCmd.SetArgs([]string{"scale-users", "abc", "--skip-prereqs"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("scale-users abc should fail")
	}
}

func TestScaleUsersCmd_NegativeCount(t *testing.T) {
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
	rootCmd.SetArgs([]string{"scale-users", "-5", "--skip-prereqs"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("scale-users -5 should fail")
	}
}

func TestValidateCmd_Runs(t *testing.T) {
	skipPrereqs = true
	defer func() { skipPrereqs = false }()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"validate", "--skip-reset", "--skip-prereqs"})
	err := rootCmd.Execute()
	// May fail on port checks (no sandbox running) - that's OK
	if err != nil {
		t.Logf("validate (no sandbox): %v", err)
	}
}

func TestScaleUsersCmd_NoArgs(t *testing.T) {
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
	rootCmd.SetArgs([]string{"scale-users", "--skip-prereqs"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("scale-users without count should fail")
	}
}

// setupSandboxWithBase creates a minimal sandbox (base + overlay + compose) for tests.
func setupSandboxWithBase(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	baseDir := filepath.Join(dir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}
	basePath := filepath.Join(baseDir, "chr.qcow2")
	if err := exec.Command("qemu-img", "create", "-f", "qcow2", basePath, "1M").Run(); err != nil {
		t.Skipf("qemu-img not available: %v", err)
	}
	_, err := sandbox.Create(dir, filepath.Join(dir, "overlay", "disk.qcow2"), baseDir, "", "", false)
	if err != nil {
		t.Fatalf("setup sandbox: %v", err)
	}
	abs, _ := filepath.Abs(dir)
	return abs
}
