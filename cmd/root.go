package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tiklab/mikrotik-sandbox/internal/config"
	"github.com/tiklab/mikrotik-sandbox/internal/logger"
	"github.com/tiklab/mikrotik-sandbox/internal/prereqs"
)

var (
	cfg          *config.Config
	verbose      bool
	skipPrereqs  bool
)

var rootCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "MikroTik Sandbox CLI - local containerized RouterOS environment for testing",
	Long: `MikroTik Sandbox CLI generates a realistic, containerized MikroTik RouterOS (CHR)
environment with synthetic users and network traffic for testing ISP billing systems,
Hotspot managers, and automation tools locally.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger.SetVerbose(verbose)
		cfg = config.Load()
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&skipPrereqs, "skip-prereqs", false, "skip prerequisite checks (for testing)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func requirePrereqs(cmd *cobra.Command, args []string) error {
	checks, err := prereqs.CheckAll()
	if err != nil {
		return fmt.Errorf("prerequisite check failed: %w", err)
	}
	if !checks.OK {
		return fmt.Errorf("missing prerequisites: %s", checks.Summary())
	}
	return nil
}
