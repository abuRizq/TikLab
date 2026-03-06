package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiklab/mikrotik-sandbox/internal/logger"
	"github.com/tiklab/mikrotik-sandbox/internal/sandbox"
)

var createForce bool

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Initialize the default isp_small synthetic environment",
	Long: `Creates a new sandbox environment with the hardcoded isp_small profile
(approx. 300 users) with DHCP and Hotspot enabled. This is the default beta mode.

Requires a CHR base image in workdir/base/: download from mikrotik.com/download/chr
and place chr.qcow2 or chr.img (or any .qcow2/.img) in the base directory.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if skipPrereqs {
			return nil
		}
		return requirePrereqs(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Verbose("create: workdir=%s profile=%s force=%v", cfg.WorkDir, cfg.Profile, createForce)
		logger.Info("Creating sandbox environment (profile: %s)...", cfg.Profile)
		res, err := sandbox.Create(cfg.WorkDir, cfg.OverlayPath, cfg.BaseDir, createForce)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		logger.Verbose("overlay=%s base=%s", res.OverlayPath, res.BasePath)
		logger.Info("Sandbox created successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().BoolVar(&createForce, "force", false, "overwrite existing sandbox")
}

