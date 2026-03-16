package cli

import (
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Provision a new sandbox environment",
		Long:  "Creates a sandbox container. Run `tiklab start` to activate it.",
		RunE:  runCreate,
	}
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Full implementation in Phase 3 (T015)
	cmd.Println("create: not yet implemented")
	return nil
}
