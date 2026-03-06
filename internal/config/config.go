package config

import (
	"os"
	"path/filepath"
)

// Config holds sandbox configuration.
type Config struct {
	WorkDir     string
	Profile     string
	BaseImage   string
	BaseDir     string
	OverlayDir  string
	OverlayPath string
	ComposePath string
	DeployDir   string
}

// Load returns the current sandbox configuration.
func Load() *Config {
	workDir := os.Getenv("SANDBOX_WORKDIR")
	if workDir == "" {
		home, _ := os.UserHomeDir()
		workDir = filepath.Join(home, ".mikrotik-sandbox")
	}
	baseDir := filepath.Join(workDir, "base")
	overlayDir := filepath.Join(workDir, "overlay")
	return &Config{
		WorkDir:     workDir,
		Profile:     "isp_small",
		BaseImage:   "chr.qcow2",
		BaseDir:     baseDir,
		OverlayDir:  overlayDir,
		OverlayPath: filepath.Join(overlayDir, "disk.qcow2"),
		ComposePath: filepath.Join(workDir, "docker-compose.yml"),
		DeployDir:   filepath.Join(workDir, "deploy"),
	}
}
