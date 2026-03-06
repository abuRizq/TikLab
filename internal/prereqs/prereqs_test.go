package prereqs

import (
	"strings"
	"testing"
)

func TestCheckAll_ReturnsResult(t *testing.T) {
	r, err := CheckAll()
	if err != nil {
		t.Fatalf("CheckAll() error = %v", err)
	}
	if r == nil {
		t.Fatal("CheckAll() returned nil result")
	}
}

func TestResult_Summary_WhenOK(t *testing.T) {
	r := &Result{OK: true}
	if s := r.Summary(); s != "all checks passed" {
		t.Errorf("Summary() = %q, want 'all checks passed'", s)
	}
}

func TestResult_Summary_WhenMissing(t *testing.T) {
	r := &Result{OK: false, Missing: []string{"docker", "qemu"}}
	s := r.Summary()
	if !strings.Contains(s, "docker") || !strings.Contains(s, "qemu") {
		t.Errorf("Summary() = %q, should contain missing items", s)
	}
}
