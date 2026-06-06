package cli

import (
	"fmt"

	"github.com/okonomipizza/statecast/internal/daemon"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the statecast daemon",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := daemon.Stop(); err != nil {
			return err
		}

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "statecast daemon stopped"); err != nil {
			return fmt.Errorf("write stop output: %w", err)
		}
		return nil
	},
}
