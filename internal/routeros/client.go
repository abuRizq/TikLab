package routeros

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-routeros/routeros"
	"github.com/tiklab/tiklab/internal/debug"
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
	host string
	port int
	user string
	pass string
}

// NewClient creates a new RouterOS API client.
func NewClient() *Client {
	return &Client{}
}

// Connect establishes a connection to the RouterOS API.
func (c *Client) Connect(host string, port int, user, pass string) error {
	c.host = host
	c.port = port
	c.user = user
	c.pass = pass
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := routeros.Dial(addr, user, pass)
	if err != nil {
		return fmt.Errorf("failed to connect to RouterOS at %s: %w", addr, err)
	}
	c.conn = conn
	return nil
}

// Reconnect closes the current connection and establishes a new one.
// Retries up to 10 times with 3s delays to tolerate transient disruption
// (e.g. hotspot enable reconfiguring the network stack in emulation mode).
func (c *Client) Reconnect() error {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	var lastErr error
	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(3 * time.Second)
		}
		// #region agent log
		debug.Log("routeros/client.go:Reconnect", "attempt", map[string]interface{}{
			"attempt": i + 1,
		}, "H1")
		// #endregion
		conn, err := dialWithTimeout(addr, c.user, c.pass, 10*time.Second)
		if err != nil {
			lastErr = err
			// #region agent log
			debug.Log("routeros/client.go:Reconnect", "attempt_failed", map[string]interface{}{
				"attempt": i + 1, "err": err.Error(),
			}, "H1")
			// #endregion
			continue
		}
		c.conn = conn
		// #region agent log
		debug.Log("routeros/client.go:Reconnect", "success", map[string]interface{}{
			"attempt": i + 1,
		}, "H1")
		// #endregion
		return nil
	}
	return fmt.Errorf("reconnect to RouterOS after 10 attempts: %w", lastErr)
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
	// #region agent log
	start := time.Now()
	debug.Log("routeros/client.go:WaitForReady", "start", map[string]interface{}{
		"host": host, "port": port, "user": user, "timeoutSec": int(timeout.Seconds()),
	}, "H2")
	// #endregion

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

		// #region agent log
		if attempt <= 3 || attempt%10 == 0 {
			sshAddr := net.JoinHostPort(host, "2222")
			sshConn, sshErr := net.DialTimeout("tcp", sshAddr, 2*time.Second)
			if sshErr == nil {
				sshBuf := make([]byte, 128)
				sshConn.SetReadDeadline(time.Now().Add(2 * time.Second))
				sn, _ := sshConn.Read(sshBuf)
				sshBanner := ""
				if sn > 0 {
					sshBanner = string(sshBuf[:sn])
				}
				sshConn.Close()
				debug.Log("routeros/client.go:WaitForReady", "ssh_probe", map[string]interface{}{
					"attempt": attempt, "elapsedSec": int(time.Since(start).Seconds()),
					"sshOpen": true, "bannerLen": sn, "banner": sshBanner,
				}, "H7")
			} else {
				debug.Log("routeros/client.go:WaitForReady", "ssh_probe", map[string]interface{}{
					"attempt": attempt, "elapsedSec": int(time.Since(start).Seconds()),
					"sshOpen": false, "sshErr": sshErr.Error(),
				}, "H7")
			}
		}

		tcpConn, tcpErr := net.DialTimeout("tcp", addr, 2*time.Second)
		if tcpErr == nil {
			apiCmd := []byte{0x06, '/', 'l', 'o', 'g', 'i', 'n', 0x00}
			tcpConn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			nw, writeErr := tcpConn.Write(apiCmd)
			tcpConn.SetReadDeadline(time.Now().Add(3 * time.Second))
			rawBuf := make([]byte, 256)
			nr, readErr := tcpConn.Read(rawBuf)
			rawHex := ""
			if nr > 0 {
				rawHex = fmt.Sprintf("%x", rawBuf[:nr])
			}
			readErrStr := ""
			if readErr != nil {
				readErrStr = readErr.Error()
			}
			writeErrStr := ""
			if writeErr != nil {
				writeErrStr = writeErr.Error()
			}
			tcpConn.Close()
			debug.Log("routeros/client.go:WaitForReady", "raw_api_probe", map[string]interface{}{
				"attempt": attempt, "elapsedSec": int(time.Since(start).Seconds()),
				"wrote": nw, "writeErr": writeErrStr,
				"readBytes": nr, "readHex": rawHex, "readErr": readErrStr,
			}, "H6")
		} else {
			debug.Log("routeros/client.go:WaitForReady", "tcp_failed", map[string]interface{}{
				"attempt": attempt, "elapsedSec": int(time.Since(start).Seconds()),
				"tcpErr": tcpErr.Error(),
			}, "H7")
		}
		// #endregion

		// #region agent log
		dialStart := time.Now()
		// #endregion
		conn, err := dialWithTimeout(addr, user, pass, 10*time.Second)
		// #region agent log
		dialMs := time.Since(dialStart).Milliseconds()
		// #endregion
		if err != nil {
			// #region agent log
			debug.Log("routeros/client.go:WaitForReady", "dial_failed", map[string]interface{}{
				"attempt": attempt, "elapsedSec": int(time.Since(start).Seconds()),
				"dialMs": dialMs, "err": err.Error(),
				"tcpWasOpen": tcpErr == nil,
			}, "H5")
			// #endregion
			log.Printf("[routeros] Boot attempt %d: %v", attempt, err)
			sleep := interval
			if tcpErr == nil && err != nil && (strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "timeout")) {
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
		c.host = host
		c.port = port
		c.user = user
		c.pass = pass
		// #region agent log
		debug.Log("routeros/client.go:WaitForReady", "success", map[string]interface{}{
			"attempt": attempt, "elapsedSec": int(time.Since(start).Seconds()),
		}, "H2")
		// #endregion
		return nil
	}
	// #region agent log
	debug.Log("routeros/client.go:WaitForReady", "timeout_exceeded", map[string]interface{}{
		"totalAttempts": attempt, "elapsedSec": int(time.Since(start).Seconds()),
		"timeoutSec":    int(timeout.Seconds()),
	}, "H2")
	// #endregion
	return fmt.Errorf("RouterOS failed to boot within timeout")
}
