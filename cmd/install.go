package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	. "github.com/little-angry-clouds/kubernetes-binaries-managers/internal/helpers"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

type osArchError struct {
	err  string
	os   string
	arch string
}

func (e *osArchError) Error() string {
	var error string
	if e.arch == "" {
		error = fmt.Sprintf("%s\nos: %s", e.err, e.os)
	} else {
		error = fmt.Sprintf("%s\narch: %s", e.err, e.arch)
	}

	return error
}

type downloadBinaryError struct {
	err  string
	url  string
	body string
}

func (e *downloadBinaryError) Error() string {
	var error string
	if e.body == "" {
		error = fmt.Sprintf("%s\nurl: %s", e.err, e.url)
	} else {
		error = fmt.Sprintf("%s\nurl: %s\nbody: %s", e.err, e.url, e.body)
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

func getOsArch() (string, error) {
	var supportedOS = []string{"linux", "windows", "darwin"}
	var supportedArch = []string{"arm", "arm64", "amd64"}
	var os = runtime.GOOS
	var arch = runtime.GOARCH

	if !contains(supportedOS, os) {
		return "", &osArchError{"os not supported", os, ""}
	}

	if !contains(supportedArch, arch) {
		return "", &osArchError{"arch not supported", "", arch}
	}

	osArch := os + "/" + arch

	return osArch, nil
}

func downloadBinary(version string) ([]byte, error) {
	var osArch string
	var err error
	var body []byte
	var errorCodeFail = 404
	var errorCodePass = 200

	// Don't control the error since at this point it should be controlled
	osArch, _ = getOsArch()
	os := strings.Split(osArch, "/")[0]
	arch := strings.Split(osArch, "/")[1]
	url := fmt.Sprintf(BinaryDownloadURL, version, os, arch)

	if strings.Contains(osArch, "windows") {
		url += windowsSuffix
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
		return body, &downloadBinaryError{"binary not found", url, string(body)}
	} else if resp.StatusCode != errorCodePass {
		return body, &downloadBinaryError{"unhandled error", url, string(body)}
	}

	if err != nil {
		return body, err
	}

	return body, nil
}

func saveBinary(fileName string, body []byte) error {
	err := ioutil.WriteFile(fileName, body, 0750)

	if err != nil {
		return err
	}

	return nil
}

func install(cmd *cobra.Command, args []string) { // nolint:funlen
	// TODO añadir soporte para fzf
	var err error
	var osArch string
	var expectedArgLength int = 1

	// TODO cambiar por ExactArgs
	if len(args) == 0 {
		fmt.Println("You must specify a version!")
		os.Exit(0)
	} else if len(args) != expectedArgLength {
		fmt.Println("Too many arguments.")
		_ = cmd.Help()
		os.Exit(0)
	}

	var version = args[0]
	// Check if os/arch is supported
	osArch, err = getOsArch()

	if err, ok := err.(*osArchError); ok {
		if err.err == "os not supported" {
			fmt.Printf("The OS '%s' is not supported.\n", err.os)
		}

		if err.err == "arch not supported" {
			fmt.Printf("The arch '%s' is not supported.\n", err.arch)
		}

		os.Exit(0)
	}
	// Set base bin directory
	home, _ := homedir.Dir()
	fileName := fmt.Sprintf("%s/.bin/%s-v%s", home, BinaryToInstall, version)

	if strings.Contains(osArch, "windows") {
		fileName += windowsSuffix
	}
	// Check if binary exists locally
	if FileExists(fileName) {
		fmt.Printf("The version %s is already installed!\n", version)
		os.Exit(0)
	}
	// Download binary
	body, err := downloadBinary(version)
	// Check for errors when downloading the binary
	if err, ok := err.(*downloadBinaryError); ok {
		if err.err == "binary not found" {
			fmt.Println("The binary was not found. The url is:")
			fmt.Println(err.url)
			os.Exit(0)
		}

		if err.err == "unhandled error" {
			fmt.Println("There was an unhandled error downloading the binary, sorry:")
			fmt.Printf("Url: %s\n", err.url)
			fmt.Printf("Error: %s\n", err.body)
		}
	}

	CheckGenericError(err)

	err = saveBinary(fileName, body)

	CheckGenericError(err)
	fmt.Printf("Done! Saving it at %s.\n", fileName)
}

func init() {
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install helm binary",
		Run:   install,
	}

	RootCmd.AddCommand(installCmd)
}
