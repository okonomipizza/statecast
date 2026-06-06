package cli

import (
	"fmt"

	"github.com/okonomipizza/statecast/internal/client"
	"github.com/spf13/cobra"
)

var registerName string

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register an agent",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if registerName == "" {
			return fmt.Errorf("agent name is required: use --name")
		}

		c, err := client.New()
		if err != nil {
			return err
		}

		agent, err := c.Register(registerName)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), agent.Name); err != nil {
			return fmt.Errorf("write register output: %w", err)
		}

		return nil
	},
}

func init() {
	registerCmd.Flags().StringVar(&registerName, "name", "", "Agent name")
	_ = registerCmd.MarkFlagRequired("name")
}
