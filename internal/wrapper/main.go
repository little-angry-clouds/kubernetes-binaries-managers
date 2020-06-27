package wrapper

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/little-angry-clouds/kubernetes-binaries-managers/internal/helpers"
	"github.com/mitchellh/go-homedir"
)

func Wrapper(binName string) { // nolint: funlen
	home, _ := homedir.Dir()
	var binPath string = fmt.Sprintf("%s/.bin", home)
	var defaultVersion string = fmt.Sprintf("%s/.%s-version", binPath, binName)
	var localVersion string = fmt.Sprintf(".%s_version", binName)
	var rawVersion []byte
	var finalVersion string
	var fileExt string
	var err error

	defaultVersion, _ = filepath.Abs(defaultVersion)
	localVersion, _ = filepath.Abs(localVersion)

	if _, err := os.Stat(localVersion); err == nil {
		rawVersion, err = ioutil.ReadFile(localVersion)
		if err != nil {
			fmt.Println("File reading error", err)
			return
		}
	} else {
		if _, err := os.Stat(defaultVersion); err != nil {
			d := []byte("auto\n")

			err = ioutil.WriteFile(defaultVersion, d, 0750)

			if err != nil {
				return
			}
		}

		rawVersion, err = ioutil.ReadFile(defaultVersion)

		if err != nil {
			fmt.Println("File reading error", err)
			return
		}
	}

	finalVersion = strings.Trim(string(rawVersion), "\n")

	if runtime.GOOS == "windows" {
		fileExt = ".exe"
	}

	if finalVersion == "auto" && binName == "kubectl" {
		version, err := helpers.KubeGetVersion()

		if err != nil {
			fmt.Println("Error getting kubernetes version: ", err)
			return
		}

		bin := fmt.Sprintf("%s/%s-v%s%s", binPath, binName, version, fileExt)
		bin, _ = filepath.Abs(bin)

		if !helpers.FileExists(bin) {
			args := []string{"install", version}
			cmd := exec.Command("kbenv"+fileExt, args...) // nolint: gosec
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()

			helpers.CheckGenericError(err)
		}

		finalVersion = version
	}

	bin := fmt.Sprintf("%s/%s-v%s", binPath, binName, finalVersion)
	bin, _ = filepath.Abs(bin)
	bin += fileExt

	cmd := exec.Command(bin, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}
