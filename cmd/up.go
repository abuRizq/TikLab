package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tiklab/mikrotik-sandbox/internal/logger"
	"github.com/tiklab/mikrotik-sandbox/internal/sandbox"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create (if needed) and start the sandbox in one command",
	Long: `Creates the sandbox environment if it does not exist, then starts it.
Equivalent to: sandbox create [--force] && sandbox start`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if skipPrereqs {
			return nil
		}
		return requirePrereqs(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Verbose("up: workdir=%s", cfg.WorkDir)
		composePath := cfg.ComposePath
		_, err := os.Stat(composePath)
		needCreate := os.IsNotExist(err) || createForce
		if needCreate {
			if createForce && err == nil {
				logger.Info("Recreating sandbox (--force)...")
			} else {
				logger.Info("Sandbox not found, creating...")
			}
			if _, createErr := sandbox.Create(cfg.WorkDir, cfg.OverlayPath, cfg.BaseDir, cfg.VolumeName, cfg.HostPath, createForce); createErr != nil {
				return fmt.Errorf("create: %w", createErr)
			}
		}
		logger.Info("Starting sandbox environment (building CHR + traffic containers, first run ~5-10 min)...")
		if err := sandbox.Start(cfg.WorkDir); err != nil {
			return fmt.Errorf("start: %w", err)
		}
		logger.Info("Sandbox started. API: localhost:8728, Winbox: localhost:8291, SSH: localhost:2222")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().BoolVar(&createForce, "force", false, "overwrite existing sandbox before recreate")
}
