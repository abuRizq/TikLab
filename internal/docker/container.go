package docker

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

const tikLabNetworkName = "tiklab-net"

// PortMapping defines host-to-container port mappings.
type PortMapping struct {
	SSH     int // Host port for SSH (container 22)
	API     int // Host port for RouterOS API (container 8728)
	Winbox  int // Host port for Winbox (container 8291)
	Control int // Host port for behavior engine control API (container 9090)
}

// DefaultPortMapping returns the default port mapping.
func DefaultPortMapping() PortMapping {
	return PortMapping{
		SSH:     2222,
		API:     8728,
		Winbox:  8291,
		Control: 9090,
	}
}

// CreateContainer creates a new container with the given name, image tag, and port mappings.
// Returns the container ID on success.
func (c *Client) CreateContainer(ctx context.Context, name, imageTag string, ports PortMapping) (containerID string, err error) {
	if c.cli == nil {
		return "", fmt.Errorf("Docker client not connected")
	}

	portBindings := nat.PortMap{
		nat.Port("22/tcp"):   {{HostPort: fmt.Sprintf("%d", ports.SSH)}},
		nat.Port("8728/tcp"): {{HostPort: fmt.Sprintf("%d", ports.API)}},
		nat.Port("8291/tcp"): {{HostPort: fmt.Sprintf("%d", ports.Winbox)}},
		nat.Port("9090/tcp"): {{HostPort: fmt.Sprintf("%d", ports.Control)}},
	}

	exposedPorts := nat.PortSet{
		nat.Port("22/tcp"):   {},
		nat.Port("8728/tcp"): {},
		nat.Port("8291/tcp"): {},
		nat.Port("9090/tcp"): {},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		CapAdd:       []string{"NET_ADMIN"},
		Resources: container.Resources{
			Devices: []container.DeviceMapping{
				{
					PathOnHost:        "/dev/net/tun",
					PathInContainer:   "/dev/net/tun",
					CgroupPermissions: "mrw",
				},
			},
		},
	}

	config := &container.Config{
		Image:        imageTag,
		ExposedPorts: exposedPorts,
	}

	resp, err := c.cli.ContainerCreate(ctx, config, hostConfig, &network.NetworkingConfig{}, nil, name)
	if err != nil {
		return "", wrapConnectionError("container create", err)
	}
	return resp.ID, nil
}

// StartContainer starts the container with the given ID.
func (c *Client) StartContainer(ctx context.Context, id string) error {
	if c.cli == nil {
		return fmt.Errorf("Docker client not connected")
	}
	if err := c.cli.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return wrapConnectionError("container start", err)
	}
	return nil
}

// StopContainer stops the container with the given ID.
func (c *Client) StopContainer(ctx context.Context, id string) error {
	if c.cli == nil {
		return fmt.Errorf("Docker client not connected")
	}
	timeout := 10 // seconds
	if err := c.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout}); err != nil {
		return wrapConnectionError("container stop", err)
	}
	return nil
}

// RemoveContainer removes the container with the given ID.
// If removeVolumes is true, associated volumes are also removed.
func (c *Client) RemoveContainer(ctx context.Context, id string, removeVolumes bool) error {
	if c.cli == nil {
		return fmt.Errorf("Docker client not connected")
	}
	if err := c.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true, RemoveVolumes: removeVolumes}); err != nil {
		return wrapConnectionError("container remove", err)
	}
	return nil
}

// ContainerLogs returns the container's stdout and stderr as a string.
func (c *Client) ContainerLogs(ctx context.Context, id string) (string, error) {
	if c.cli == nil {
		return "", fmt.Errorf("Docker client not connected")
	}
	opts := container.LogsOptions{ShowStdout: true, ShowStderr: true}
	rdr, err := c.cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		return "", err
	}
	defer rdr.Close()
	var out, errOut bytes.Buffer
	_, _ = stdcopy.StdCopy(&out, &errOut, rdr)
	return out.String() + errOut.String(), nil
}

// ContainerInspect returns the container inspection data (state, exit code, etc.).
func (c *Client) ContainerInspect(ctx context.Context, id string) (status string, exitCode int, err error) {
	if c.cli == nil {
		return "", 0, fmt.Errorf("Docker client not connected")
	}
	inspect, err := c.cli.ContainerInspect(ctx, id)
	if err != nil {
		return "", 0, err
	}
	return inspect.State.Status, inspect.State.ExitCode, nil
}

// ContainerExists returns true if a container with the given name exists.
func (c *Client) ContainerExists(ctx context.Context, name string) (bool, error) {
	if c.cli == nil {
		return false, fmt.Errorf("Docker client not connected")
	}
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, wrapConnectionError("container list", err)
	}
	for _, cnt := range containers {
		for _, n := range cnt.Names {
			if n == "/"+name || n == name {
				return true, nil
			}
		}
	}
	return false, nil
}

// EnsureNetwork creates the tiklab-net bridge network if it doesn't already exist.
// The second network interface (eth1) is required by the RouterOS base image's
// entrypoint to enable QEMU user-mode networking with hostfwd port forwarding.
// Without it, QEMU only uses TAP networking and ports are not forwarded to the VM.
func (c *Client) EnsureNetwork(ctx context.Context) (string, error) {
	if c.cli == nil {
		return "", fmt.Errorf("Docker client not connected")
	}
	networks, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return "", wrapConnectionError("network list", err)
	}
	for _, n := range networks {
		if n.Name == tikLabNetworkName {
			return n.ID, nil
		}
	}
	resp, err := c.cli.NetworkCreate(ctx, tikLabNetworkName, network.CreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return "", wrapConnectionError("network create", err)
	}
	return resp.ID, nil
}

// ConnectToNetwork attaches a container to the given network.
// Silently succeeds if the container is already connected.
func (c *Client) ConnectToNetwork(ctx context.Context, containerID, networkID string) error {
	if c.cli == nil {
		return fmt.Errorf("Docker client not connected")
	}
	err := c.cli.NetworkConnect(ctx, networkID, containerID, &network.EndpointSettings{})
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return nil
	}
	return err
}
