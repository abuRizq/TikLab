package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tiklab/mikrotik-sandbox/internal/logger"
	"github.com/tiklab/mikrotik-sandbox/internal/sandbox"
)

var scaleUsersCmd = &cobra.Command{
	Use:   "scale-users [count]",
	Short: "Dynamically add or remove active simulation namespaces",
	Long: `Scales the number of active simulated users (network namespaces) up or down
without rebuilding the full environment. Pass the target count.`,
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if n, err := strconv.Atoi(args[0]); err != nil || n < 0 {
			return fmt.Errorf("count must be a non-negative integer, got %q", args[0])
		}
		if skipPrereqs {
			return nil
		}
		return requirePrereqs(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		count, _ := strconv.Atoi(args[0])
		logger.Verbose("scale-users: target=%d workdir=%s", count, cfg.WorkDir)
		logger.Info("Scaling users to %d...", count)
		if err := sandbox.ScaleUsers(cfg.WorkDir, count); err != nil {
			return fmt.Errorf("scale-users: %w", err)
		}
		logger.Info("User count scaled to %d.", count)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scaleUsersCmd)
}
