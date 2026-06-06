package cli

import (
	"github.com/spf13/cobra"
)

// rootCmd は statecast CLI のルートコマンド。
var rootCmd = &cobra.Command{
	Use:   "statecast",
	Short: "Share state between AI agents",
}

// Execute はルートコマンドを実行する。
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(startCmd, stopCmd, registerCmd, listCmd, updateCmd, getCmd)
}
