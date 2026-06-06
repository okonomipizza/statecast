package cli

import (
	"fmt"

	"github.com/okonomipizza/statecast/internal/client"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get agent state",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, err := requireAgentName(cmd)
		if err != nil {
			return err
		}

		c, err := client.New()
		if err != nil {
			return err
		}

		state, err := c.GetState(name)
		if err != nil {
			return err
		}

		_, err = cmd.OutOrStdout().Write(state)
		if err != nil {
			return fmt.Errorf("write state to stdout: %w", err)
		}

		return nil
	},
}

func init() {
	getCmd.Flags().String("name", "", "Agent name")
	_ = getCmd.MarkFlagRequired("name")
}
