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
	// VolumeName is set when running in Docker; used for compose volume binding
	VolumeName string
	// HostPath is set via TIKLAB_HOST_PATH for bind-mount Docker mode
	HostPath string
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
	volName := os.Getenv("TIKLAB_VOLUME")
	hostPath := os.Getenv("TIKLAB_HOST_PATH")
	if hostPath == "" {
		hostPath = os.Getenv("TIKLAB_DATA")
	}
	if volName == "" && workDir == "/sandbox" && hostPath == "" {
		volName = "tiklab-data"
	}
	return &Config{
		WorkDir:     workDir,
		Profile:     "isp_small",
		BaseImage:   "chr.qcow2",
		BaseDir:     baseDir,
		OverlayDir:  overlayDir,
		OverlayPath: filepath.Join(overlayDir, "disk.qcow2"),
		ComposePath: filepath.Join(workDir, "docker-compose.yml"),
		DeployDir:   filepath.Join(workDir, "deploy"),
		VolumeName:  volName,
		HostPath:    hostPath,
	}
}
