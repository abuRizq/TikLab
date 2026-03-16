package cli

import (
	"github.com/spf13/cobra"
)

func newScaleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scale [count]",
		Short: "Adjust the number of simulated users",
		Long:  "Scales the active simulated user count (1-500). Sandbox must be running.",
		Args:  cobra.ExactArgs(1),
		RunE:  runScale,
	}
}

func runScale(cmd *cobra.Command, args []string) error {
	// Full implementation in Phase 6 (T037)
	cmd.Println("scale: not yet implemented")
	return nil
}
