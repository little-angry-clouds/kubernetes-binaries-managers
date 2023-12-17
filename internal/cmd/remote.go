package cmd

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	. "github.com/nixknight/binaries-managers/internal/helpers"
	. "github.com/nixknight/binaries-managers/internal/versions"
	"github.com/spf13/cobra"
)

func remote(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		fmt.Println("Too many arguments.")

		_ = cmd.Help()

		os.Exit(0)
	}
	// TODO meter soporte para fzf
	var versions []*version.Version
	var err error
	var allReleases bool
	var allVersions bool

	versions, err = GetRemoteVersions(VersionsAPI)
	CheckGenericError(err)
	allReleases, err = cmd.Flags().GetBool("all-releases")
	CheckGenericError(err)
	allVersions, err = cmd.Flags().GetBool("all-versions")
	CheckGenericError(err)
	versions, err = SortVersions(versions, allReleases, allVersions)
	CheckGenericError(err)
	PrintVersions(versions)
}

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "List remote versions",
	Run:   remote,
}

func init() {
	listCmd.AddCommand(remoteCmd)
}
