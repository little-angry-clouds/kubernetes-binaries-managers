package cmd

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	. "github.com/nixknight/binaries-managers/internal/helpers"
	. "github.com/nixknight/binaries-managers/internal/versions"
	"github.com/spf13/cobra"
)

func local(cmd *cobra.Command, args []string) {
	var err error
	var allReleases bool
	var allVersions bool
	var versions []*version.Version

	if len(args) != 0 {
		fmt.Println("Too many arguments.")

		_ = cmd.Help()

		os.Exit(0)
	}

	allReleases, err = cmd.Flags().GetBool("all-releases")

	CheckGenericError(err)

	allVersions, err = cmd.Flags().GetBool("all-versions")

	CheckGenericError(err)

	versions, err = GetLocalVersions(BinaryToInstall)

	CheckGenericError(err)

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
