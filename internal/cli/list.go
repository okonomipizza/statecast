package cli

import (
	"fmt"

	"github.com/okonomipizza/statecast/internal/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered agents",
	RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		agents, err := c.List()
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		if _, err := fmt.Fprintln(out, "# name"); err != nil {
			return fmt.Errorf("write list output: %w", err)
		}
		for _, agent := range agents {
			if _, err := fmt.Fprintln(out, agent.Name); err != nil {
				return fmt.Errorf("write list output: %w", err)
			}
		}

		return nil
	},
}
