package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/helpers"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func use(cmd *cobra.Command, args []string) {
	var err error
	var entryScript string
	var minArgsLength = 1

	if len(args) == 0 {
		fmt.Println("You must specify a version!")
		os.Exit(0)
	} else if len(args) != minArgsLength {
		fmt.Println("Too many arguments.")
		_ = cmd.Help()
		os.Exit(0)
	}

	var version = args[0]

	if runtime.GOOS == "windows" {
		entryScript = `
$BIN_PATH="%s/%s-v"
$DEFAULT_VERSION="$($BIN_PATH)%s"
$LOCAL=".%s_version"
$FILE_EXISTS=Test-Path $LOCAL
if ($FILE_EXISTS -eq $True) {
	local_version=cat $LOCAL
	$version="$($BIN_PATH)$($local_version)"
}
else {
	$version="$DEFAULT_VERSION"
}
Invoke-Expression "$version $args"
`
	} else {
		entryScript = `
#!/bin/sh
BIN_PATH="%s/%s-v"
DEFAULT_VERSION="${BIN_PATH}%s"
LOCAL=".%s_version"
if [ -e "$LOCAL" ]
then
	version="${BIN_PATH}$(cat $LOCAL)"
else
	version="${DEFAULT_VERSION}"
fi
eval $version $@
`
	}

	home, _ := homedir.Dir()
	binPath := fmt.Sprintf("%s/.bin", home)
	script := []byte(fmt.Sprintf(entryScript, binPath, BinaryToInstall, version, BinaryToInstall))
	defaultBin := fmt.Sprintf("%s/%s", binPath, BinaryToInstall)

	if runtime.GOOS == "windows" {
		defaultBin += ".ps1"
	}

	err = ioutil.WriteFile(defaultBin, script, 0750)
	CheckGenericError(err)

	fmt.Printf("Done! Using %s version.\n", version)
}

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Set the default version to use",
	Run:   use,
}

func init() {
	RootCmd.AddCommand(useCmd)
}
