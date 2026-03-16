package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tiklab/tiklab/internal/docker"
	"github.com/tiklab/tiklab/internal/routeros"
	"github.com/tiklab/tiklab/internal/sandbox"
)

const routerOSBootTimeout = 90 * time.Second

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Activate the sandbox and start traffic generation",
		Long:  "Starts the sandbox container, boots RouterOS, applies configuration, and begins traffic generation.",
		RunE:  runStart,
	}
}

func runStart(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Connect to Docker
	dc := docker.NewClient()
	if err := dc.Connect(); err != nil {
		return err
	}
	defer dc.Close()

	if !dc.IsAvailable() {
		return fmt.Errorf("Docker is not running. Please start Docker and try again")
	}

	// Build wait-for-ready function using RouterOS client
	user, pass := routeros.DefaultCredentials()
	waitForReady := func(ctx context.Context, host string, port int) error {
		cmd.Println("Waiting for RouterOS to boot...")
		ros := routeros.NewClient()
		defer ros.Close()
		return ros.WaitForReady(host, port, user, pass, routerOSBootTimeout)
	}

	cmd.Println("Starting sandbox...")
	mgr := sandbox.NewManager(dc)
	if err := mgr.Start(ctx, waitForReady); err != nil {
		return err
	}

	// Apply initial configuration (DHCP, Hotspot, queues)
	state, err := sandbox.Load()
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("No sandbox found. Run `tiklab create` first")
	}

	ros := routeros.NewClient()
	defer ros.Close()
	if err := ros.Connect("127.0.0.1", state.Ports.API, user, pass); err != nil {
		return fmt.Errorf("failed to connect to RouterOS for config: %w", err)
	}

	if err := routeros.ApplyInitialConfig(ros, cmd.Println); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	cmd.Println("Starting traffic generation (50 users)...")
	if err := callEngineStart(state.Ports.Control, 50); err != nil {
		return fmt.Errorf("failed to start behavior engine: %w", err)
	}

	// Update state UserCount
	state.UserCount = 50
	if err := sandbox.Save(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	cmd.Println("Ready.")
	cmd.Println()
	cmd.Println("  SSH:    ssh admin@localhost -p 2222")
	cmd.Println("  API:    localhost:8728")
	cmd.Println("  Winbox: localhost:8291")
	return nil
}
