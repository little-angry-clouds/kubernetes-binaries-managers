package helpers

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"os/exec"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func KubeGetVersion() (string, error) {
	var kubeconfig string
	var cli *string
	var err error
	var home string
	var version string
	var config *rest.Config

	cli = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	// Try first --kubeconfig parameter
	if *cli != "" {
		kubeconfig = *cli
	} else {
		home, _ = homedir.Dir()

		kubeconfigVar := os.Getenv("KUBECONFIG")

		// Try first KUBECONFIG env var
		// If not set, try default kubeconfig path
		// If not find, let it empty
		if kubeconfigVar != "" { // nolint: gocritic
			kubeconfig = kubeconfigVar
		} else if FileExists(filepath.Join(home, ".kube", "config")) {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// If no kubeconfig
	if kubeconfig == "" {
		version = getDefaultVersion()
		return version, nil
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	config.Timeout = 1 * time.Second

	client, err := kubernetes.NewForConfig(config)

	if err != nil {
		version = getDefaultVersion()
		return version, nil
	}

	v, err := client.DiscoveryClient.ServerVersion()
	if err != nil {
		version = getDefaultVersion()
		return version, nil
	}

	version = v.String()[1:]

	return version, nil
}

func getDefaultVersion() string {
	var fileExt string
	var version string

	if runtime.GOOS == "windows" {
		fileExt = ".exe"
	}

	// See if there's any kubectl version installed
	args := []string{"list", "local"}
	cmd := exec.Command("kbenv"+fileExt, args...) // nolint: gosec
	output, _ := cmd.Output()
	out := string(output)

	// If there's no kubectl, get latest
	// If there's at least one, use that
	if out == "" {
		resp, err := http.Get("https://storage.googleapis.com/kubernetes-release/release/stable.txt")
		CheckGenericError(err)

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		CheckGenericError(err)

		bodyText := string(body)
		version = strings.Trim(bodyText[1:], "\n")
	} else {
		lines := strings.Split(out, "\n")
		version = lines[0]
	}

	return version
}
