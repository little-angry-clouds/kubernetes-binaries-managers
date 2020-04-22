package versions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
)

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
