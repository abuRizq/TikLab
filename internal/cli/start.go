package cli

import (
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Activate the sandbox and start traffic generation",
		Long:  "Starts the sandbox container, boots RouterOS, applies configuration, and begins traffic generation.",
		RunE:  runStart,
	}
}

func runStart(cmd *cobra.Command, args []string) error {
	// Full implementation in Phase 3 (T016)
	cmd.Println("start: not yet implemented")
	return nil
}
