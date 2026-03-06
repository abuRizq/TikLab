package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiklab/mikrotik-sandbox/internal/logger"
	"github.com/tiklab/mikrotik-sandbox/internal/sandbox"
)

var destroyAll bool

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Tear down the environment and clean up",
	Long: `Completely tears down the sandbox environment: stops containers,
removes namespaces, bridges, and temporary assets. Use --all to also remove
the workdir (base image and all state).`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if skipPrereqs {
			return nil
		}
		return requirePrereqs(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Verbose("destroy: workdir=%s all=%v", cfg.WorkDir, destroyAll)
		logger.Info("Destroying sandbox environment...")
		if err := sandbox.Destroy(cfg.WorkDir, destroyAll); err != nil {
			return fmt.Errorf("destroy: %w", err)
		}
		logger.Info("Sandbox destroyed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVar(&destroyAll, "all", false, "also remove workdir (base image and all state)")
}
