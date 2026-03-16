package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tiklab/tiklab/internal/docker"
	"github.com/tiklab/tiklab/internal/sandbox"
)

func newDestroyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "destroy",
		Short: "Remove the sandbox completely",
		Long:  "Stops the container and removes all Docker artifacts.",
		RunE:  runDestroy,
	}
}

func runDestroy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Connect to Docker
	dc := docker.NewClient()
	if err := dc.Connect(); err != nil {
		return err
	}
	defer dc.Close()

	cmd.Println("Destroying sandbox...")
	mgr := sandbox.NewManager(dc)
	if err := mgr.Destroy(ctx); err != nil {
		return err
	}
	cmd.Println("Sandbox destroyed.")
	return nil
}
