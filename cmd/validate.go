package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tiklab/mikrotik-sandbox/internal/logger"
	"github.com/tiklab/mikrotik-sandbox/internal/sandbox"
	"github.com/tiklab/mikrotik-sandbox/internal/validate"
)

var validateSkipReset bool

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Run beta validation checks against the sandbox",
	Long: `Runs validation checks: port connectivity (API, Winbox, SSH),
optional API CRUD test, and reset performance. Requires a running sandbox.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if skipPrereqs {
			return nil
		}
		return requirePrereqs(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		host := "127.0.0.1"
		timeout := 5 * time.Second

		// 1. Port checks
		logger.Info("Checking ports %s:%d, %d, %d...", host, validate.APIPort, validate.WinboxPort, validate.SSHPort)
		portResults := validate.CheckAllPorts(host, timeout)
		allPortsOK := true
		for _, r := range portResults {
			if r.Open {
				logger.Info("  %s (port %d): OK (latency %v)", r.Name, r.Port, r.Latency.Round(time.Millisecond))
			} else {
				logger.Error("  %s (port %d): FAIL - %v", r.Name, r.Port, r.Err)
				allPortsOK = false
			}
		}

		// 2. API CRUD (optional, requires CHR with API enabled)
		logger.Info("Checking API CRUD...")
		apiCfg := validate.DefaultAPIConfig()
		apiResult := validate.CheckAPI(apiCfg)
		if apiResult.Connected && apiResult.Err == nil {
			logger.Info("  API CRUD: OK")
		} else {
			logger.Verbose("  API CRUD: %v (CHR may not be ready or API disabled)", apiResult.Err)
		}

		// 3. Reset performance (optional, requires created sandbox; performs actual reset)
		if cfg != nil && !validateSkipReset {
			logger.Info("Measuring reset performance (target < 2s)...")
			result := validate.MeasureReset(func() error {
				return sandbox.Reset(cfg.WorkDir, cfg.OverlayPath, cfg.BaseDir)
			})
			if result.Err != nil {
				logger.Verbose("  Reset: %v (sandbox may not be created)", result.Err)
			} else if result.OK {
				logger.Info("  Reset: OK (%v)", result.Duration.Round(time.Millisecond))
			} else {
				logger.Error("  Reset: SLOW (%v, target < 2s)", result.Duration.Round(time.Millisecond))
			}
		}

		if !allPortsOK {
			return fmt.Errorf("port checks failed")
		}
		logger.Info("Validation complete.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&validateSkipReset, "skip-reset", false, "skip reset performance measurement")
}
