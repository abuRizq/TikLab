package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiklab/tiklab/internal/routeros"
	"github.com/tiklab/tiklab/internal/sandbox"
)

func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset sandbox to clean state",
		Long:  "Wipes all configuration changes and regenerates simulated users with fresh identities.",
		RunE:  runReset,
	}
}

func runReset(cmd *cobra.Command, args []string) error {
	state, err := sandbox.Load()
	if err != nil {
		return err
	}
	if state == nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: No sandbox found. Run `tiklab create` first.")
		return fmt.Errorf("no sandbox found")
	}
	if state.Status != sandbox.StatusRunning {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: Sandbox is not running. Run `tiklab start` first.")
		return fmt.Errorf("sandbox is not running")
	}

	cmd.Println("Resetting sandbox...")

	cmd.Println("Clearing configuration...")
	if err := callEngineStop(state.Ports.Control); err != nil {
		return fmt.Errorf("failed to stop behavior engine: %w", err)
	}

	ros := routeros.NewClient()
	defer ros.Close()
	user, pass := routeros.DefaultCredentials()
	if err := ros.Connect("127.0.0.1", state.Ports.API, user, pass); err != nil {
		return fmt.Errorf("failed to connect to RouterOS: %w", err)
	}
	if err := routeros.WipeConfig(ros); err != nil {
		return fmt.Errorf("failed to wipe configuration: %w", err)
	}

	cmd.Println("Reapplying initial setup...")
	if err := routeros.ApplyInitialConfig(ros, func(msg string) { cmd.Println(msg) }); err != nil {
		return fmt.Errorf("failed to reapply configuration: %w", err)
	}

	cmd.Println("Regenerating users (50 users)...")
	if err := callEngineStart(state.Ports.Control, 50); err != nil {
		return fmt.Errorf("failed to start behavior engine: %w", err)
	}

	state.UserCount = 50
	if err := sandbox.Save(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	cmd.Println("Reset complete.")
	return nil
}
