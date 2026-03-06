package validate

import (
	"fmt"
	"net"
	"time"
)

// Ports holds the sandbox service ports.
const (
	APIPort    = 8728
	WinboxPort = 8291
	SSHPort    = 2222
)

// PortResult holds the result of a port check.
type PortResult struct {
	Port   int
	Name   string
	Open   bool
	Err    error
	Latency time.Duration
}

// CheckPort attempts a TCP connection to host:port.
func CheckPort(host string, port int, timeout time.Duration) PortResult {
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	latency := time.Since(start)
	if err != nil {
		return PortResult{Port: port, Open: false, Err: err, Latency: latency}
	}
	conn.Close()
	return PortResult{Port: port, Open: true, Latency: latency}
}

// CheckAllPorts verifies API, Winbox, and SSH ports on host.
func CheckAllPorts(host string, timeout time.Duration) []PortResult {
	if host == "" {
		host = "127.0.0.1"
	}
	ports := []struct {
		port int
		name string
	}{
		{APIPort, "API"},
		{WinboxPort, "Winbox"},
		{SSHPort, "SSH"},
	}
	results := make([]PortResult, len(ports))
	for i, p := range ports {
		r := CheckPort(host, p.port, timeout)
		r.Name = p.name
		results[i] = r
	}
	return results
}
