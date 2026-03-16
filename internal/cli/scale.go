package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tiklab/tiklab/internal/sandbox"
)

const (
	minUserCount = 1
	maxUserCount = 500
)

func newScaleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scale [count]",
		Short: "Adjust the number of simulated users",
		Long:  "Scales the active simulated user count (1-500). Sandbox must be running.",
		Args:  cobra.ExactArgs(1),
		RunE:  runScale,
	}
}

func runScale(cmd *cobra.Command, args []string) error {
	count, err := strconv.Atoi(args[0])
	if err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: Invalid user count. Provide a number between 1 and 500.")
		return err
	}
	if count < minUserCount {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: Minimum user count is 1.")
		return fmt.Errorf("minimum user count is 1")
	}
	if count > maxUserCount {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: Maximum user count is 500.")
		return fmt.Errorf("maximum user count is 500")
	}

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

	cmd.Println("Scaling to", count, "users...")
	if err := callEngineScale(state.Ports.Control, count); err != nil {
		return fmt.Errorf("failed to scale: %w", err)
	}

	state.UserCount = count
	if err := sandbox.Save(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	cmd.Println("Scaled to", count, "users.")
	return nil
}
