package validate

import (
	"testing"
	"time"
)

func TestDefaultAPIConfig(t *testing.T) {
	cfg := DefaultAPIConfig()
	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host = %q", cfg.Host)
	}
	if cfg.Port != APIPort {
		t.Errorf("Port = %d", cfg.Port)
	}
	if cfg.User != "admin" {
		t.Errorf("User = %q", cfg.User)
	}
}

func TestCheckAPI_NoServer(t *testing.T) {
	// No CHR running - should fail to connect
	cfg := DefaultAPIConfig()
	cfg.Timeout = 100 * time.Millisecond
	result := CheckAPI(cfg)
	if result.Connected && result.Err == nil {
		t.Error("expected failure when no API server")
	}
}
