package main

import (
	"os"

	"github.com/little-angry-clouds/kubernetes-binaries-managers/internal/cmd"
	"github.com/mitchellh/go-homedir"
)

var binaryDownloadURL string = "https://storage.googleapis.com/kubernetes-release/release/v%s/bin/%s/%s/kubectl" // nolint:lll
var versionsAPI string = "https://api.github.com/repos/kubernetes/kubernetes/releases?per_page=100&page="        // nolint:lll

func main() {
	home, _ := homedir.Dir()
	_ = os.MkdirAll(home+"/.bin", os.ModePerm)
	cmd.BinaryDownloadURL = binaryDownloadURL
	cmd.VersionsAPI = versionsAPI
	cmd.BinaryToInstall = "kubectl"
	cmd.RootCmd.Use = "kbenv"
	cmd.RootCmd.Short = "Kubectl version manager"
	cmd.Execute()
}
