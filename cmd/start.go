package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiklab/mikrotik-sandbox/internal/logger"
	"github.com/tiklab/mikrotik-sandbox/internal/sandbox"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Boot the Docker Compose stack, QEMU CHR, and Traffic Engine",
	Long: `Starts the sandbox environment: Docker Compose stack, QEMU CHR router,
and the traffic engine. Requires a prior 'sandbox create'.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if skipPrereqs {
			return nil
		}
		return requirePrereqs(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Verbose("start: workdir=%s", cfg.WorkDir)
		logger.Info("Starting sandbox environment...")
		if err := sandbox.Start(cfg.WorkDir); err != nil {
			return fmt.Errorf("start: %w", err)
		}
		logger.Info("Sandbox started. API: localhost:8728, Winbox: localhost:8291, SSH: localhost:2222")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
