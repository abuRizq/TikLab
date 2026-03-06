package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiklab/mikrotik-sandbox/internal/logger"
	"github.com/tiklab/mikrotik-sandbox/internal/sandbox"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Revert CHR to post-create state via QCOW2 overlay drop",
	Long: `Instantly reverts the CHR to its post-create state by dropping and
recreating the QCOW2 overlay disk. Target: < 2 seconds.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if skipPrereqs {
			return nil
		}
		return requirePrereqs(cmd, args)
	},
		RunE: func(cmd *cobra.Command, args []string) error {
		logger.Verbose("reset: workdir=%s overlay=%s", cfg.WorkDir, cfg.OverlayPath)
		logger.Info("Resetting sandbox to post-create state...")
		if err := sandbox.Reset(cfg.WorkDir, cfg.OverlayPath, cfg.BaseDir); err != nil {
			return fmt.Errorf("reset: %w", err)
		}
		logger.Info("Sandbox reset complete.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
