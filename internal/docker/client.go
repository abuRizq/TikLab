package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

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
		return fmt.Errorf("failed to create Docker client: %w", err)
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

// PullImage pulls the given image tag and streams progress to stdout.
// The tag should be in the form "repository:tag" (e.g., "tiklab/sandbox:0.1.0").
func (c *Client) PullImage(ctx context.Context, tag string, out io.Writer) error {
	if c.cli == nil {
		return fmt.Errorf("Docker client not connected")
	}
	reader, err := c.cli.ImagePull(ctx, tag, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", tag, err)
	}
	defer reader.Close()

	if out != nil {
		_, err = io.Copy(out, reader)
		if err != nil {
			return fmt.Errorf("failed to stream pull output: %w", err)
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
