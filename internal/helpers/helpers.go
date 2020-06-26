package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// CheckGenericError checks if there's an error, shows it and exits the program if it is
func CheckGenericError(err error) {
	if err != nil {
		message := fmt.Sprintf("An error was detected, exiting: %s", err)
		fmt.Fprintf(os.Stderr, "%s\n", message)
		os.Exit(1) // nolint:gomnd
	}
}

func CheckHTTPError(resp *http.Response) {
	var result map[string]interface{}
	var message string

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)

		if resp.Header.Get("Content-Type") == "application/json" {
			CheckGenericError(err)
			err = json.Unmarshal(body, &result)
			CheckGenericError(err)

			message = result["message"].(string)
		} else {
			message = string(body)
		}

		fmt.Fprintf(os.Stderr, "An error detected getting all versions: %s", message)
		os.Exit(1) // nolint:gomnd
	}
}

// https://golangcode.com/check-if-a-file-exists/
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func GetLastPage(link string) (int, error) {
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

type OSArchError struct {
	Err  string
	OS   string
	Arch string
}

func (e *OSArchError) Error() string {
	var error string
	if e.Arch == "" {
		error = fmt.Sprintf("%s\nos: %s", e.Err, e.OS)
	} else {
		error = fmt.Sprintf("%s\narch: %s", e.Err, e.Arch)
	}

	return error
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}

	return false
}

func GetOSArch() (string, error) {
	var supportedOS = []string{"linux", "windows", "darwin"}
	var supportedArch = []string{"arm", "arm64", "amd64"}
	var os = runtime.GOOS
	var arch = runtime.GOARCH

	if !contains(supportedOS, os) {
		return "", &OSArchError{"os not supported", os, ""}
	}

	if !contains(supportedArch, arch) {
		return "", &OSArchError{"arch not supported", "", arch}
	}

	osArch := os + "/" + arch

	return osArch, nil
}
