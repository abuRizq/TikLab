package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tiklab/tiklab/internal/docker"
	"github.com/tiklab/tiklab/internal/ports"
	"github.com/tiklab/tiklab/internal/sandbox"
)

const (
	defaultImageTag      = "tiklab/sandbox:0.1.0"
	defaultContainerName = "tiklab-sandbox"
)

func getImageTag() string {
	if tag := os.Getenv("TIKLAB_IMAGE"); tag != "" {
		return tag
	}
	return defaultImageTag
}

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Provision a new sandbox environment",
		Long:  "Creates a sandbox container. Run `tiklab start` to activate it.",
		RunE:  runCreate,
	}
}

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Connect to Docker
	dc := docker.NewClient()
	if err := dc.Connect(); err != nil {
		return fmt.Errorf("Docker is not running. Please start Docker and try again")
	}
	defer dc.Close()

	// Pre-flight: Docker daemon reachable
	if !dc.IsAvailable() {
		return fmt.Errorf("Docker is not running. Please start Docker and try again")
	}

	// Pre-flight: port conflicts
	portsToCheck := []int{2222, 8728, 8291, 9090}
	for _, p := range portsToCheck {
		if ports.InUse(p) {
			return fmt.Errorf("Port %d is already in use. Free the port or configure an alternative", p)
		}
	}

	// Check no sandbox exists
	state, err := sandbox.Load()
	if err != nil {
		return err
	}
	if state != nil {
		return fmt.Errorf("Sandbox already exists. Run `tiklab destroy` first")
	}

	// Pull image if needed
	imageTag := getImageTag()
	exists, err := dc.ImageExists(ctx, imageTag)
	if err != nil {
		return err
	}
	if !exists {
		cmd.Println("Pulling image", imageTag+"...")
		if err := dc.PullImage(ctx, imageTag, os.Stdout); err != nil {
			return fmt.Errorf("failed to pull image: %w. Run 'make build-image' to build locally", err)
		}
	}

	// Create sandbox
	cmd.Println("Creating sandbox...")
	mgr := sandbox.NewManager(dc)
	if err := mgr.Create(ctx, imageTag, defaultContainerName); err != nil {
		return err
	}

	cmd.Println("Sandbox created.")
	cmd.Println()
	cmd.Println("  SSH:    localhost:2222")
	cmd.Println("  API:    localhost:8728")
	cmd.Println("  Winbox: localhost:8291")
	cmd.Println()
	cmd.Println("Run `tiklab start` to activate.")
	return nil
}
