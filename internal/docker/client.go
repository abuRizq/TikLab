package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

// wrapConnectionError wraps Docker connectivity errors with a user-friendly message.
// Detects connection refused, reset, EOF, and socket errors during operations.
func wrapConnectionError(op string, err error) error {
	if err == nil {
		return nil
	}
	if !isConnectionError(err) {
		return err
	}
	return fmt.Errorf("Docker connection lost. Is Docker running? (%s: %w)", op, err)
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	// syscall.ECONNREFUSED, ECONNRESET
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.ECONNREFUSED || errno == syscall.ECONNRESET
	}
	// *net.OpError (wraps dial/read/write failures)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	// Common error substrings from Docker client
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "connection reset") ||
		strings.Contains(s, "dial tcp") ||
		strings.Contains(s, "eof") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "no such host")
}

// Client wraps the Docker API client with TikLab-specific operations.
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client. Call Connect() before use.
func NewClient() *Client {
	return &Client{}
}

// Connect establishes a connection to the Docker daemon using environment variables
// and API version negotiation.
func (c *Client) Connect() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return wrapConnectionError("connect", err)
	}
	c.cli = cli
	return nil
}

// Close releases resources held by the client.
func (c *Client) Close() error {
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}

// IsAvailable returns true if the Docker daemon is reachable.
func (c *Client) IsAvailable() bool {
	if c.cli == nil {
		return false
	}
	ctx := context.Background()
	_, err := c.cli.Ping(ctx)
	return err == nil
}

// ImageExists returns true if the given image tag exists locally.
func (c *Client) ImageExists(ctx context.Context, tag string) (bool, error) {
	if c.cli == nil {
		return false, fmt.Errorf("Docker client not connected")
	}
	_, _, err := c.cli.ImageInspectWithRaw(ctx, tag)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, wrapConnectionError("image inspect", err)
	}
	return true, nil
}

// PullImage pulls the given image tag and streams progress to stdout.
// The tag should be in the form "repository:tag" (e.g., "tiklab/sandbox:0.1.0").
func (c *Client) PullImage(ctx context.Context, tag string, out io.Writer) error {
	if c.cli == nil {
		return fmt.Errorf("Docker client not connected")
	}
	reader, err := c.cli.ImagePull(ctx, tag, types.ImagePullOptions{})
	if err != nil {
		return wrapConnectionError("image pull", err)
	}
	defer reader.Close()

	if out != nil {
		_, err = io.Copy(out, reader)
		if err != nil {
			return wrapConnectionError("image pull stream", err)
		}
	} else {
		_, _ = io.Copy(io.Discard, reader)
	}
	return nil
}

// Raw returns the underlying Docker client for advanced operations.
func (c *Client) Raw() *client.Client {
	return c.cli
}
