package wrapper

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/go-homedir"
)

func Wrapper(binName string) {
	home, _ := homedir.Dir()
	var binPath string = fmt.Sprintf("%s/.bin", home)
	var defaultVersion string = fmt.Sprintf("%s/.%s-version", binPath, binName)
	var localVersion string = fmt.Sprintf(".%s_version", binName)
	var rawVersion []byte
	var finalVersion string
	var err error

	if _, err := os.Stat(localVersion); err == nil {
		rawVersion, err = ioutil.ReadFile(localVersion)
		if err != nil {
			fmt.Println("File reading error", err)
			return
		}
	} else {
		rawVersion, err = ioutil.ReadFile(defaultVersion)
		if err != nil {
			fmt.Println("File reading error", err)
			return
		}
	}

	finalVersion = strings.Trim(string(rawVersion), "\n")

	cmd := exec.Command(fmt.Sprintf("%s-v%s", binName, finalVersion), os.Args[1:]...) // nolint: gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if strings.Contains(err.Error(), "executable file not found in $PATH") {
		fmt.Printf("%s\n", err)
	}
}
