package cli

import (
	"github.com/okonomipizza/statecast/internal/client"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [json]",
	Short: "Update agent state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := requireAgentName(cmd)
		if err != nil {
			return err
		}

		c, err := client.New()
		if err != nil {
			return err
		}

		return c.UpdateState(name, []byte(args[0]))
	},
}

func init() {
	updateCmd.Flags().String("name", "", "Agent name")
	_ = updateCmd.MarkFlagRequired("name")
}
