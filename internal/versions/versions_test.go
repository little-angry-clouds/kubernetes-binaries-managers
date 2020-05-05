package versions

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/helpers"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

func TestSortVersions(t *testing.T) { // nolint: funlen
	var flagtests = []struct {
		testName       string
		input          []string
		output         []string
		allReleases    bool
		allVersions    bool
		versionsLength int
	}{
		{
			"stable releases",
			[]string{
				"1.17.5", "1.18.2", "1.16.9", "1.18.1", "1.18.0",
				"1.15.11", "1.17.4", "1.16.8", "1.15.10", "1.16.7",
				"1.17.3", "1.15.9", "1.17.2", "1.16.6", "1.16.0-alpha.1",
				"1.3.0-alpha.3", "1.6.0-beta.3", "1.13.5", "1.14.0-beta.1",
				"1.13.0-alpha.1", "1.7.10", "1.15.1", "1.10.0-beta.2", "1.17.0-alpha.2",
				"1.7.13", "1.9.2", "1.8.7", "1.9.1", "1.7.12", "1.8.6",
			},
			[]string{
				"1.18.2", "1.18.1", "1.18.0", "1.17.5", "1.17.4",
				"1.17.3", "1.17.2", "1.16.9", "1.16.8", "1.16.7",
				"1.16.6", "1.15.11", "1.15.10", "1.15.9", "1.15.1",
				"1.13.5", "1.9.2", "1.9.1", "1.8.7", "1.8.6",
			},
			false,
			false,
			20,
		},
		{
			"mixed releases",
			[]string{
				"1.17.5", "1.18.2", "1.16.9", "1.18.1", "1.18.0",
				"1.15.11", "1.17.4", "1.16.8", "1.15.10", "1.16.7",
				"1.17.3", "1.15.9", "1.17.2", "1.16.6", "1.16.0-alpha.1",
				"1.3.0-alpha.3", "1.6.0-beta.3", "1.13.5", "1.14.0-beta.1",
				"1.13.0-alpha.1", "1.7.10", "1.15.1", "1.10.0-beta.2", "1.17.0-alpha.2",
				"1.7.13", "1.9.2", "1.8.7", "1.9.1", "1.7.12", "1.8.6",
			},
			[]string{
				"1.18.2", "1.18.1", "1.18.0", "1.17.5", "1.17.4",
				"1.17.3", "1.17.2", "1.17.0-alpha.2", "1.16.9", "1.16.8",
				"1.16.7", "1.16.6", "1.16.0-alpha.1", "1.15.11", "1.15.10",
				"1.15.9", "1.15.1", "1.14.0-beta.1", "1.13.5", "1.13.0-alpha.1",
			},
			true,
			false,
			20,
		},
		{
			"all non mixed releases",
			[]string{
				"1.17.5", "1.18.2", "1.16.9", "1.18.1", "1.18.0",
				"1.15.11", "1.17.4", "1.16.8", "1.15.10", "1.16.7",
				"1.17.3", "1.15.9", "1.17.2", "1.16.6", "1.16.0-alpha.1",
				"1.3.0-alpha.3", "1.6.0-beta.3", "1.13.5", "1.14.0-beta.1",
				"1.13.0-alpha.1", "1.7.10", "1.15.1", "1.10.0-beta.2", "1.17.0-alpha.2",
				"1.7.13", "1.9.2", "1.8.7", "1.9.1", "1.7.12", "1.8.6",
			},
			[]string{
				"1.18.2", "1.18.1", "1.18.0", "1.17.5", "1.17.4",
				"1.17.3", "1.17.2", "1.16.9", "1.16.8", "1.16.7",
				"1.16.6", "1.15.11", "1.15.10", "1.15.9", "1.15.1",
				"1.13.5", "1.9.2", "1.9.1", "1.8.7", "1.8.6",
				"1.7.13", "1.7.12", "1.7.10",
			},
			false,
			true,
			23,
		},
		{
			"all mixed releases",
			[]string{
				"1.17.5", "1.18.2", "1.16.9", "1.18.1", "1.18.0",
				"1.15.11", "1.17.4", "1.16.8", "1.15.10", "1.16.7",
				"1.17.3", "1.15.9", "1.17.2", "1.16.6", "1.16.0-alpha.1",
				"1.3.0-alpha.3", "1.6.0-beta.3", "1.13.5", "1.14.0-beta.1",
				"1.13.0-alpha.1", "1.7.10", "1.15.1", "1.10.0-beta.2", "1.17.0-alpha.2",
				"1.7.13", "1.9.2", "1.8.7", "1.9.1", "1.7.12", "1.8.6",
			},
			[]string{
				"1.18.2", "1.18.1", "1.18.0", "1.17.5", "1.17.4",
				"1.17.3", "1.17.2", "1.17.0-alpha.2", "1.16.9", "1.16.8",
				"1.16.7", "1.16.6", "1.16.0-alpha.1", "1.15.11", "1.15.10",
				"1.15.9", "1.15.1", "1.14.0-beta.1", "1.13.5", "1.13.0-alpha.1",
				"1.10.0-beta.2", "1.9.2", "1.9.1", "1.8.7", "1.8.6",
				"1.7.13", "1.7.12", "1.7.10", "1.6.0-beta.3", "1.3.0-alpha.3",
			},
			true,
			true,
			30,
		},
	}

	for _, tt := range flagtests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			expectedVersionsStr := tt.output
			receivedVersionsStr := tt.input
			receivedVersionsVrs := make([]*version.Version, len(tt.input))
			for i, raw := range receivedVersionsStr {
				v, _ := version.NewVersion(raw)
				receivedVersionsVrs[i] = v
			}
			actualVersionsVrs, err := SortVersions(receivedVersionsVrs, tt.allReleases, tt.allVersions)
			actualVersionsStr := make([]string, len(actualVersionsVrs))
			for i, raw := range actualVersionsVrs {
				v := fmt.Sprintf("%v", raw)
				actualVersionsStr[i] = v
			}
			assert.Nil(t, err)
			assert.Equal(t, expectedVersionsStr, actualVersionsStr)
			assert.Equal(t, tt.versionsLength, len(actualVersionsStr))
		})
	}
}

func TestGetLastPage(t *testing.T) {
	var flagtests = []struct {
		testName string
		input    string
		expected int
	}{
		{"value 20", "<https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"next\", <https://api.github.com/repositories/20580498/releases?per_page=20&page=20>; rel=\"last\"", 20}, // nolint: lll
		{"value 30", "<https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"next\", <https://api.github.com/repositories/20580498/releases?per_page=20&page=30>; rel=\"last\"", 30}, // nolint: lll
		{"value 2", "<https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"next\", <https://api.github.com/repositories/20580498/releases?per_page=20&page=0>; rel=\"last\"", 2},    // nolint: lll
		{"vaue 1", "<https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"next\", <https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"last\"", 1},     // nolint: lll
	}

	for _, tt := range flagtests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			actualLastPage, err := GetLastPage(tt.input)
			expectedLastPage := tt.expected
			assert.Nil(t, err)
			assert.Equal(t, expectedLastPage, actualLastPage)
		})
	}
}

func TestGetRemoteVersions(t *testing.T) { // nolint: funlen
	var flagtests = []struct {
		testName string
		input    string
		expected []string
	}{
		{
			"one page",
			"2",
			[]string{
				"1.17.5", "1.18.2", "1.16.9", "1.18.1", "1.19.0-alpha.1", "1.18.0", "1.18.0-rc.1", "1.15.11", "1.17.4", "1.16.8",
				"1.18.0-beta.2", "1.18.0-beta.1", "1.18.0-alpha.5", "1.15.10", "1.16.7",
				"1.17.3", "1.18.0-alpha.3", "1.15.9", "1.17.2", "1.16.6",
			},
		},
		{
			"two pages",
			"3",
			[]string{
				"1.17.5", "1.18.2", "1.16.9", "1.18.1", "1.19.0-alpha.1",
				"1.18.0", "1.18.0-rc.1", "1.15.11", "1.17.4", "1.16.8",
				"1.18.0-beta.2", "1.18.0-beta.1", "1.18.0-alpha.5", "1.15.10", "1.16.7",
				"1.17.3", "1.18.0-alpha.3", "1.15.9", "1.17.2", "1.16.6",
				"1.18.0-alpha.2", "1.15.8", "1.16.5", "1.17.1", "1.18.0-alpha.1",
				"1.14.10", "1.16.4", "1.15.7", "1.17.0", "1.17.0-rc.2",
				"1.17.0-rc.1", "1.17.0-beta.2", "1.16.3", "1.14.9", "1.15.6",
				"1.17.0-beta.1", "1.17.0-alpha.3", "1.15.5", "1.16.2", "1.13.12",
			},
		},
	}

	for _, tt := range flagtests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				page := req.URL.Query()["page"][0]
				fakeResponse := fmt.Sprintf("test_data/page_%s", page)
				jsonFile, err := os.Open(fakeResponse)
				if err != nil {
					return
				}
				defer jsonFile.Close()
				jsonBytes, err := ioutil.ReadAll(jsonFile)
				if err != nil {
					return
				}
				link := fmt.Sprintf(
					"<https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"next\", <https://api.github.com/repositories/20580498/releases?per_page=20&page=%s>; rel=\"last\"", // nolint: lll
					tt.input)
				rw.Header().Set("Content-Type", "application/json")
				rw.Header().Set("Link", link)
				_, err = rw.Write(jsonBytes)
				if err != nil {
					return
				}
			}))
			defer server.Close()

			r, err := GetRemoteVersions(server.URL + "/?page=")
			assert.Nil(t, err)
			expectedVersions := tt.expected
			receivedVersions := make([]string, len(r))
			for i, raw := range r {
				v := fmt.Sprintf("%v", raw)
				receivedVersions[i] = v
			}
			assert.Equal(t, expectedVersions, receivedVersions)
		})
	}
}

func TestGetLocalVersions(t *testing.T) {
	var flagtests = []struct {
		testName string
		input    []string
		expected []string
	}{
		{
			"all releases",
			[]string{
				"1.17.5", "1.18.2", "1.16.9", "1.18.1", "1.18.0",
				"1.15.11", "1.17.4", "1.16.8", "1.15.10", "1.16.7",
				"1.17.3", "1.15.9", "1.17.2", "1.16.6", "1.16.0-alpha.1",
				"1.3.0-alpha.3", "1.6.0-beta.3", "1.13.5", "1.14.0-beta.1",
				"1.13.0-alpha.1", "1.7.10", "1.15.1", "1.10.0-beta.2", "1.17.0-alpha.2",
				"1.7.13", "1.9.2", "1.8.7", "1.9.1", "1.7.12", "1.8.6",
			},
			[]string{
				"1.10.0-beta.2", "1.13.0-alpha.1", "1.13.5", "1.14.0-beta.1", "1.15.1",
				"1.15.10", "1.15.11", "1.15.9", "1.16.0-alpha.1", "1.16.6",
				"1.16.7", "1.16.8", "1.16.9", "1.17.0-alpha.2", "1.17.2",
				"1.17.3", "1.17.4", "1.17.5", "1.18.0", "1.18.1",
				"1.18.2", "1.3.0-alpha.3", "1.6.0-beta.3", "1.7.10", "1.7.12",
				"1.7.13", "1.8.6", "1.8.7", "1.9.1", "1.9.2",
			},
		},
	}

	for _, tt := range flagtests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			expectedVersions := tt.expected
			receivedVersions := tt.input
			binaryName := "binaryTest"
			home, _ := homedir.Dir()
			for _, value := range receivedVersions {
				binary := fmt.Sprintf("%s/.bin/%s-v%s", home, binaryName, value)
				_, _ = os.Create(binary)
				defer os.Remove(binary)
			}
			actualVersionsVrs, err := GetLocalVersions(binaryName)
			actualVersionsStr := make([]string, len(actualVersionsVrs))
			for i, raw := range actualVersionsVrs {
				v := fmt.Sprintf("%v", raw)
				actualVersionsStr[i] = v
			}
			assert.Nil(t, err)
			assert.Equal(t, expectedVersions, actualVersionsStr)
		})
	}
}
