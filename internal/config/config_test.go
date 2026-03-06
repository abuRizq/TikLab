package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultWorkDir(t *testing.T) {
	os.Unsetenv("SANDBOX_WORKDIR")
	cfg := Load()
	if cfg.WorkDir == "" {
		t.Error("WorkDir should not be empty")
	}
	if !filepath.HasSuffix(cfg.WorkDir, ".mikrotik-sandbox") {
		t.Errorf("WorkDir = %q, should end with .mikrotik-sandbox", cfg.WorkDir)
	}
}

func TestLoad_CustomWorkDir(t *testing.T) {
	os.Setenv("SANDBOX_WORKDIR", "/custom/path")
	defer os.Unsetenv("SANDBOX_WORKDIR")
	cfg := Load()
	if cfg.WorkDir != "/custom/path" {
		t.Errorf("WorkDir = %q, want /custom/path", cfg.WorkDir)
	}
}

func TestLoad_Profile(t *testing.T) {
	cfg := Load()
	if cfg.Profile != "isp_small" {
		t.Errorf("Profile = %q, want isp_small", cfg.Profile)
	}
}

func TestLoad_OverlayDir(t *testing.T) {
	cfg := Load()
	if cfg.OverlayDir == "" {
		t.Error("OverlayDir should not be empty")
	}
	if filepath.Base(cfg.OverlayDir) != "overlay" {
		t.Errorf("OverlayDir should end with overlay, got %q", cfg.OverlayDir)
	}
}
