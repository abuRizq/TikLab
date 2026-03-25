package routeros

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-routeros/routeros"
)

// dialWithTimeout runs routeros.Dial with a timeout to avoid long blocks when RouterOS is not ready.
func dialWithTimeout(addr, user, pass string, timeout time.Duration) (*routeros.Client, error) {
	type result struct {
		conn *routeros.Client
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		c, err := routeros.Dial(addr, user, pass)
		ch <- result{conn: c, err: err}
	}()
	select {
	case r := <-ch:
		return r.conn, r.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout after %v", timeout)
	}
}

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
		c.conn.Close()
	}
	return nil
}

// Run executes a RouterOS API command and returns the reply.
func (c *Client) Run(command string, args ...string) (*routeros.Reply, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to RouterOS")
	}
	return c.conn.Run(append([]string{command}, args...)...)
}

// DefaultCredentials returns the default RouterOS CHR credentials (admin, empty password).
func DefaultCredentials() (user, pass string) {
	return "admin", ""
}

// WaitForReady polls the RouterOS API until it responds or the timeout is exceeded.
// Polls every 3 seconds.
func (c *Client) WaitForReady(host string, port int, user, pass string, timeout time.Duration) error {
	start := time.Now()
	deadline := time.Now().Add(timeout)
	interval := 3 * time.Second
	attempt := 0
	lastProgress := time.Now()

	for time.Now().Before(deadline) {
		attempt++
		addr := net.JoinHostPort(host, strconv.Itoa(port))

		if time.Since(lastProgress) >= 30*time.Second {
			elapsed := int(time.Since(start).Seconds())
			remaining := int(timeout.Seconds()) - elapsed
			log.Printf("[routeros] Still waiting... (attempt %d, %ds elapsed, %ds remaining)", attempt, elapsed, remaining)
			lastProgress = time.Now()
		}

		tcpConn, tcpErr := net.DialTimeout("tcp", addr, 2*time.Second)
		if tcpErr == nil {
			tcpConn.Close()
		}

		conn, err := dialWithTimeout(addr, user, pass, 10*time.Second)
		if err != nil {
			log.Printf("[routeros] Boot attempt %d: %v", attempt, err)
			sleep := interval
			if tcpErr == nil && (strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "timeout")) {
				sleep = 5 * time.Second
			}
			time.Sleep(sleep)
			continue
		}
		conn.Close()

		if c.conn != nil {
			c.conn.Close()
		}
		conn2, err := dialWithTimeout(addr, user, pass, 10*time.Second)
		if err != nil {
			log.Printf("[routeros] Boot attempt %d (connect): %v", attempt, err)
			sleep := interval
			if tcpErr == nil && (strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "timeout")) {
				sleep = 5 * time.Second
			}
			time.Sleep(sleep)
			continue
		}
		c.conn = conn2
		return nil
	}
	return fmt.Errorf("RouterOS failed to boot within timeout")
}
