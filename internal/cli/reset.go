package cli

import (
	"github.com/spf13/cobra"
)

func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset sandbox to clean state",
		Long:  "Wipes all configuration changes and regenerates simulated users with fresh identities.",
		RunE:  runReset,
	}
}

func runReset(cmd *cobra.Command, args []string) error {
	// Full implementation in Phase 7 (T042)
	cmd.Println("reset: not yet implemented")
	return nil
}
