package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/helpers"
	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/versions"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func local(cmd *cobra.Command, args []string) {
	var err error
	var allReleases bool
	var allVersions bool

	if len(args) != 0 {
		fmt.Println("Too many arguments.")

		_ = cmd.Help()

		os.Exit(0)
	}

	var versions []*version.Version // nolint:prealloc

	home, _ := homedir.Dir()

	binDir := fmt.Sprintf("%s/.bin/%s-v*", home, BinaryToInstall)
	matches, _ := filepath.Glob(binDir)
	allReleases, err = cmd.Flags().GetBool("all-releases")

	CheckGenericError(err)

	allVersions, err = cmd.Flags().GetBool("all-versions")

	CheckGenericError(err)

	for _, match := range matches {
		v := strings.Split(match, "/")
		vs := strings.Replace(v[len(v)-1], BinaryToInstall+"-v", "", 1)
		ver, err := version.NewVersion(vs)

		CheckGenericError(err)

		versions = append(versions, ver)
	}

	versions, err = SortVersions(versions, allReleases, allVersions)

	CheckGenericError(err)

	PrintVersions(versions)
}

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "List installed versions",
	Run:   local,
}

func init() {
	listCmd.AddCommand(localCmd)
}
