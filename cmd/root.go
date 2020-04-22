package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var BinaryDownloadURL string
var VersionsAPI string
var RootCmd = &cobra.Command{}
var BinaryToInstall string
var windowsSuffix string = ".exe"

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {}
