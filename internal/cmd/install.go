package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/binary"
	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/helpers"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func install(cmd *cobra.Command, args []string) { // nolint:funlen
	// TODO añadir soporte para fzf
	var err error
	var osArch string
	var expectedArgLength int = 1

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
	// Check if os/arch is supported
	osArch, err = GetOSArch()

	if err, ok := err.(*OSArchError); ok {
		if err.Err == "os not supported" {
			fmt.Printf("The OS '%s' is not supported.\n", err.OS)
		}

		if err.Err == "arch not supported" {
			fmt.Printf("The arch '%s' is not supported.\n", err.Arch)
		}

		os.Exit(0)
	}
	// Set base bin directory
	home, _ := homedir.Dir()
	fileName := fmt.Sprintf("%s/.bin/%s-v%s", home, BinaryToInstall, version)
	fileName, _ = filepath.Abs(fileName)

	if strings.Contains(osArch, "windows") {
		fileName += windowsSuffix
	}
	// Check if binary exists locally
	if FileExists(fileName) {
		fmt.Printf("The version %s is already installed!\n", version)
		os.Exit(0)
	}
	// Download binary
	body, err := DownloadBinary(version, BinaryDownloadURL)
	// Check for errors when downloading the binary
	if err, ok := err.(*DownloadBinaryError); ok {
		if err.Err == "binary not found" {
			fmt.Println("The binary was not found. The url is:")
			fmt.Println(err.URL)
			os.Exit(0)
		}

		if err.Err == "unhandled error" {
			fmt.Println("There was an unhandled error downloading the binary, sorry:")
			fmt.Printf("Url: %s\n", err.URL)
			fmt.Printf("Error: %s\n", err.Body)
		}
	}

	CheckGenericError(err)

	err = SaveBinary(fileName, body)

	CheckGenericError(err)
	fmt.Printf("Done! Saving it at %s.\n", fileName)
}

func init() {
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install binary",
		Run:   install,
	}

	RootCmd.AddCommand(installCmd)
}
