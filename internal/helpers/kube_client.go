package helpers

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func KubeGetVersion() (string, error) {
	var kubeconfig *string
	var err error
	var home string
	var version string

	home, _ = homedir.Dir()

	kubeconfigVar := os.Getenv("KUBECONFIG")

	if kubeconfigVar != "" { // nolint: gocritic
		kubeconfig = flag.String("kubeconfig", kubeconfigVar, "absolute path to the kubeconfig file")
	} else if FileExists(filepath.Join(home, ".kube", "config")) {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		version = getLatestVersion()
		return version, nil
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		version = getLatestVersion()
		return version, nil
	}

	v, err := client.DiscoveryClient.ServerVersion()
	if err != nil {
		version = getLatestVersion()
		return version, nil
	}

	version = v.String()[1:]

	return version, nil
}

func getLatestVersion() string {
	resp, err := http.Get("https://storage.googleapis.com/kubernetes-release/release/stable.txt")
	CheckGenericError(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	CheckGenericError(err)

	bodyText := string(body)
	version := bodyText[1:]

	version = strings.Trim(version, "\n")

	return version
}
