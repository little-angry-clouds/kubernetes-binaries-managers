package binary

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/helpers"
	"github.com/mholt/archiver/v3"
)

type DownloadBinaryError struct {
	Err  string
	URL  string
	Body string
}

func (e *DownloadBinaryError) Error() string {
	var error string
	if e.Body == "" {
		error = fmt.Sprintf("%s\nurl: %s", e.Err, e.URL)
	} else {
		error = fmt.Sprintf("%s\nurl: %s\nbody: %s", e.Err, e.URL, e.Body)
	}

	return error
}

func DownloadBinary(version string, url string) ([]byte, error) {
	var osArch string
	var err error
	var body []byte
	var errorCodeFail = 404
	var errorCodePass = 200

	// Don't control the error since at this point it should be controlled
	osArch, _ = GetOSArch()
	os := strings.Split(osArch, "/")[0]
	arch := strings.Split(osArch, "/")[1]
	url = fmt.Sprintf(url, version, os, arch)

	if strings.Contains(url, "helm") {
		if strings.Contains(osArch, "windows") {
			url += ".zip"
		} else {
			url += ".tar.gz"
		}
	} else {
		if strings.Contains(osArch, "windows") {
			url += ".exe"
		}
	}

	fmt.Println("Downloading binary...")
	resp, err := http.Get(url) // nolint

	if err != nil {
		return body, err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return body, err
	}

	if resp.StatusCode == errorCodeFail {
		return body, &DownloadBinaryError{"binary not found", url, string(body)}
	} else if resp.StatusCode != errorCodePass {
		return body, &DownloadBinaryError{"unhandled error", url, string(body)}
	}

	if err != nil {
		return body, err
	}

	return body, nil
}

func SaveBinary(fileName string, body []byte) error { // nolint: funlen
	var err error

	// helm returns a compressed file, so save it somewhere and decompress it
	if strings.Contains(fileName, "helm") {
		var fileExt string

		rand.Seed(time.Now().UnixNano())

		randomNumbers := 5000
		tempDir, err := ioutil.TempDir("", "helm")

		if err != nil {
			return err
		}
		// clean temp dir
		defer os.RemoveAll(tempDir)

		osArch, _ := GetOSArch()
		file := fmt.Sprintf("%s/helm-%s", tempDir, strconv.Itoa(rand.Intn(randomNumbers)))
		file, _ = filepath.Abs(file)

		if strings.Contains(osArch, "windows") {
			fileExt = ".zip"
		} else {
			fileExt = ".tar.gz"
		}

		err = ioutil.WriteFile(file+fileExt, body, 0750)

		if err != nil {
			return err
		}

		err = archiver.Unarchive(file+fileExt, file)

		if err != nil {
			return err
		}

		OS := strings.Split(osArch, "/")[0]
		arch := strings.Split(osArch, "/")[1]
		path, _ := filepath.Abs(file + fmt.Sprintf("/%s-%s/helm", OS, arch))

		if strings.Contains(osArch, "windows") {
			path += ".exe"
		}

		body, err = ioutil.ReadFile(path)

		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(fileName, body, 0750)

	if err != nil {
		return err
	}

	return nil
}
