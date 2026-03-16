package cli

import (
	"github.com/spf13/cobra"
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
	// Full implementation in Phase 3 (T017)
	cmd.Println("destroy: not yet implemented")
	return nil
}
