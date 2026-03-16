package routeros

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/go-routeros/routeros"
)

// Client wraps the RouterOS API client.
type Client struct {
	conn *routeros.Client
}

// NewClient creates a new RouterOS API client.
func NewClient() *Client {
	return &Client{}
}

// Connect establishes a connection to the RouterOS API.
func (c *Client) Connect(host string, port int, user, pass string) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := routeros.Dial(addr, user, pass)
	if err != nil {
		return fmt.Errorf("failed to connect to RouterOS at %s: %w", addr, err)
	}
	c.conn = conn
	return nil
}

// Close closes the connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Run executes a RouterOS API command and returns the reply.
func (c *Client) Run(command string, args ...string) (*routeros.Reply, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to RouterOS")
	}
	return c.conn.Run(command, args...)
}

// DefaultCredentials returns the default RouterOS CHR credentials (admin, empty password).
func DefaultCredentials() (user, pass string) {
	return "admin", ""
}

// WaitForReady polls the RouterOS API until it responds or the timeout is exceeded.
// Polls every 3 seconds.
func (c *Client) WaitForReady(host string, port int, user, pass string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	interval := 3 * time.Second
	attempt := 0

	for time.Now().Before(deadline) {
		attempt++
		conn, err := routeros.Dial(net.JoinHostPort(host, strconv.Itoa(port)), user, pass)
		if err != nil {
			log.Printf("[routeros] Boot attempt %d: %v", attempt, err)
			time.Sleep(interval)
			continue
		}
		_ = conn.Close()

		// Connect and keep the connection for the manager
		if c.conn != nil {
			_ = c.conn.Close()
		}
		conn2, err := routeros.Dial(net.JoinHostPort(host, strconv.Itoa(port)), user, pass)
		if err != nil {
			log.Printf("[routeros] Boot attempt %d (connect): %v", attempt, err)
			time.Sleep(interval)
			continue
		}
		c.conn = conn2
		return nil
	}
	return fmt.Errorf("RouterOS failed to boot within timeout")
}
