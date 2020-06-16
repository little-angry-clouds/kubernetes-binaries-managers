package wrapper

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

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
	bin := fmt.Sprintf("%s/%s-v%s", binPath, binName, finalVersion)
	args := append([]string{bin}, os.Args[1:]...)
	err = syscall.Exec(bin, args, os.Environ()) // golint: nosec

	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
