package helpers

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func KubeGetVersion() (string, error) {
	var kubeconfig *string
	var err error

	kubeconfigVar := os.Getenv("KUBECONFIG")
	if kubeconfigVar != "" {
		kubeconfig = flag.String("kubeconfig", kubeconfigVar, "absolute path to the kubeconfig file")
	} else if home, _ := homedir.Dir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return "", err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	v, err := client.DiscoveryClient.ServerVersion()
	if err != nil {
		return "", err
	}

	version := v.String()[1:]

	return version, nil
}
