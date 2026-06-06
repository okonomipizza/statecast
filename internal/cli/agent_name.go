package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// requireAgentName は --name フラグが必須のコマンド用に agent name を返す。
func requireAgentName(cmd *cobra.Command) (string, error) {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", fmt.Errorf("agent name is required: use --name")
	}
	return name, nil
}
