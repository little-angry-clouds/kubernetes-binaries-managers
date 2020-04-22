package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/helpers"
	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/versions"
	"github.com/spf13/cobra"
)

type Page struct {
	Release string `json:"tag_name"`
}

func getLastPage(link string) (int, error) {
	var lastPageInt int
	var err error

	link = strings.Split(link, " ")[2]
	lastPageIndex := strings.LastIndex(link, "page=")
	lastPageStr := strings.Replace(link[lastPageIndex+5:], ">;", "", 2)
	lastPageInt, err = strconv.Atoi(lastPageStr)

	if err != nil {
		return 0, err
	}

	if lastPageInt == 0 {
		lastPageInt = 2
	}

	return lastPageInt, nil
}

func getVersions(endpoint string) ([]*version.Version, error) {
	var versions []*version.Version
	var defaultHTTPTimeout time.Duration = 5
	var client = http.Client{Timeout: time.Second * defaultHTTPTimeout}

	resp, err := client.Get(endpoint + "1")
	if err != nil {
		return versions, err
	}

	defer resp.Body.Close()
	lastPage, err := getLastPage(resp.Header.Get("Link"))

	if err != nil {
		return versions, err
	}

	for page := 2; page <= lastPage; page++ {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return versions, err
		}

		rel := []Page{}

		err = json.Unmarshal(body, &rel)

		if err != nil {
			return versions, err
		}

		if page != lastPage {
			resp, err = client.Get(endpoint + strconv.Itoa(page))
			if err != nil {
				return versions, err
			}
			defer resp.Body.Close()
		}

		for _, element := range rel {
			v, err := version.NewVersion(element.Release)
			if err != nil {
				return versions, err
			}

			versions = append(versions, v)
		}
	}

	return versions, nil
}

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

	versions, err = getVersions(VersionsAPI)
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
