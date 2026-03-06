package validate

import (
	"fmt"
	"time"

	"github.com/go-routeros/routeros/v3"
)

// APIConfig holds connection parameters for the MikroTik API.
type APIConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  time.Duration
}

// DefaultAPIConfig returns config for local sandbox (admin, empty password).
func DefaultAPIConfig() APIConfig {
	return APIConfig{
		Host:     "127.0.0.1",
		Port:     APIPort,
		User:     "admin",
		Password: "",
		Timeout:  5 * time.Second,
	}
}

// APIResult holds the result of an API check.
type APIResult struct {
	Connected bool
	Err       error
}

// CheckAPI connects and runs a simple read to verify API CRUD compatibility.
func CheckAPI(cfg APIConfig) APIResult {
	if cfg.Port == 0 {
		cfg.Port = APIPort
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := routeros.Dial(addr, cfg.User, cfg.Password)
	if err != nil {
		return APIResult{Connected: false, Err: err}
	}
	defer client.Close()

	// Simple read: /system/resource print
	_, err = client.Run("/system/resource/print")
	if err != nil {
		return APIResult{Connected: true, Err: err}
	}
	return APIResult{Connected: true}
}
