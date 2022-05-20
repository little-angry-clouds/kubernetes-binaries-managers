package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	goVersion "go.hein.dev/go-version"
)

var (
	Version    = "dev"
	Commit     = "none"
	Date       = "unknown"
	shortened  = false
	output     = "json"
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Outputs the current build information",
		Run: func(_ *cobra.Command, _ []string) {
			resp := goVersion.FuncWithOutput(shortened, Version, Commit, Date, output)
			fmt.Print(resp)
		},
	}
)

func init() {
	versionCmd.Flags().BoolVar(&shortened, "short", false, "Print just the version number.")
	versionCmd.Flags().StringVar(&output, "output", "yaml", "Output format. One of 'yaml' or 'json'.")
	RootCmd.AddCommand(versionCmd)
}
