package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root tiklab command with version and help flags.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tiklab",
		Short: "TikLab Sandbox - MikroTik RouterOS simulation environment",
		Long:  "TikLab provisions and manages a Docker-based MikroTik RouterOS sandbox for development and testing.",
	}

	cmd.PersistentFlags().BoolP("help", "h", false, "Help for tiklab")
	cmd.Version = version

	// Register subcommands (stubs for Phase 2, full implementation in Phase 3+)
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newStartCmd())
	cmd.AddCommand(newScaleCmd())
	cmd.AddCommand(newResetCmd())
	cmd.AddCommand(newDestroyCmd())

	return cmd
}
