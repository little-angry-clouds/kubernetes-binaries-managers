package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists local and remote versions",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
		}
	},
}

// listCmd represents the list command
func init() {
	listCmd.PersistentFlags().Bool("all-releases", false, "return all releases, including alpha, beta and rc releases")
	listCmd.PersistentFlags().Bool("all-versions", false, "return all versions")
	RootCmd.AddCommand(listCmd)
}
