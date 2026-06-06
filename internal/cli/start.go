package cli

import (
	"fmt"

	"github.com/okonomipizza/statecast/internal/daemon"
	"github.com/okonomipizza/statecast/internal/store"
	"github.com/spf13/cobra"
)

var startForeground bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the statecast daemon",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if startForeground {
			server := daemon.NewServer(store.New())
			return server.Serve()
		}

		if err := daemon.StartBackground(); err != nil {
			return err
		}

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "statecast daemon started"); err != nil {
			return fmt.Errorf("write start output: %w", err)
		}
		return nil
	},
}

func init() {
	startCmd.Flags().BoolVar(&startForeground, "foreground", false, "Run daemon in foreground (internal use)")
	_ = startCmd.Flags().MarkHidden("foreground")
}
