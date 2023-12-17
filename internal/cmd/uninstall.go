package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
	. "github.com/nixknight/binaries-managers/internal/helpers"
	"github.com/spf13/cobra"
)

var customUninstallDir string

func uninstall(cmd *cobra.Command, args []string) {
	// TODO añadir soporte para fzf
	var err error
	var expectedArgLength int = 1
	var dirPath string

	// TODO cambiar por ExactArgs
	if len(args) == 0 {
		fmt.Println("You must specify a version!")
		os.Exit(0)
	} else if len(args) != expectedArgLength {
		fmt.Println("Too many arguments.")
		_ = cmd.Help()
		os.Exit(0)
	}

	var version = args[0]

	// Set base bin directory
	if customUninstallDir != "" {
		dirPath = customUninstallDir
	} else {
		home, _ := homedir.Dir()
		dirPath = filepath.Join(home, ".bin")
	}

	fileName := fmt.Sprintf("%s/%s-v%s", dirPath, BinaryToInstall, version)
	fileName, _ = filepath.Abs(fileName)

	if runtime.GOOS == "windows" {
		fileName += windowsSuffix
	}

	// Check if binary exists locally
	if FileExists(fileName) {
		err = os.Remove(fileName)
		CheckGenericError(err)
		fmt.Printf("Done! %s version uninstalled from %s.\n", version, fileName)
		os.Exit(0)
	}

	fmt.Printf("The version %s was already uninstalled! Doing nothing.\n", version)
}

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall binary",
	Run:   uninstall,
}

func init() {
	uninstallCmd.Flags().StringVarP(&customUninstallDir, "dir", "d", "", "custom directory path to uninstall the binary")
	RootCmd.AddCommand(uninstallCmd)
}
