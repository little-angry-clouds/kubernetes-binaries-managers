package versions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/helpers"
	"github.com/mitchellh/go-homedir"
)

type Page struct {
	Release string `json:"tag_name"`
}

func SortVersions(versions []*version.Version, allReleases bool, allVersions bool) ([]*version.Version, error) {
	var numberOfVersion int
	var finalVersions []*version.Version

	sort.Sort(sort.Reverse(version.Collection(versions)))

	if allVersions {
		numberOfVersion = len(versions) - 1
	} else {
		numberOfVersion = 19
	}

	for i := 0; i <= numberOfVersion; i++ {
		if i == len(versions) {
			break
		}

		if allReleases {
			finalVersions = append(finalVersions, versions[i])
		} else {
			if !strings.ContainsAny(versions[i].String(), "beta") &&
				!strings.ContainsAny(versions[i].String(), "alpha") &&
				!strings.ContainsAny(versions[i].String(), "rc") {
				finalVersions = append(finalVersions, versions[i])
			} else {
				versions = append(versions[:i], versions[i+1:]...)
				i--
			}
		}
	}

	return finalVersions, nil
}

func PrintVersions(versions []*version.Version) {
	for _, element := range versions {
		fmt.Println(element)
	}
}

func GetLocalVersions(binary string) ([]*version.Version, error) {
	var versions []*version.Version // nolint:prealloc

	home, _ := homedir.Dir()
	binDir := fmt.Sprintf("%s/.bin/%s-v*", home, binary)
	matches, _ := filepath.Glob(binDir)

	for _, match := range matches {
		v := strings.Split(match, "/")
		vs := strings.Replace(v[len(v)-1], binary+"-v", "", 1)
		ver, err := version.NewVersion(vs)

		if err != nil {
			return versions, err
		}

		versions = append(versions, ver)
	}

	return versions, nil
}

func GetRemoteVersions(endpoint string) ([]*version.Version, error) {
	var versions []*version.Version
	var defaultHTTPTimeout time.Duration = time.Second * 10
	var client = http.Client{Timeout: defaultHTTPTimeout}

	resp, err := client.Get(endpoint + "1")
	if err != nil {
		return versions, err
	}

	defer resp.Body.Close()
	lastPage, err := GetLastPage(resp.Header.Get("Link"))

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
