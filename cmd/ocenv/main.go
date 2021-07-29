package main

import (
	"os"

	"github.com/little-angry-clouds/kubernetes-binaries-managers/internal/cmd"
	"github.com/mitchellh/go-homedir"
)

var binaryDownloadURL string = "https://github.com/openshift/okd/releases/download/%s/openshift-client-%s-%s.tar.gz" // nolint:lll
var versionsAPI string = "https://api.github.com/repos/openshift/okd/releases?per_page=100&page="                    // nolint:lll

func main() {
	home, _ := homedir.Dir()
	_ = os.MkdirAll(home+"/.bin", os.ModePerm)
	cmd.BinaryDownloadURL = binaryDownloadURL
	cmd.VersionsAPI = versionsAPI
	cmd.BinaryToInstall = "oc"
	cmd.RootCmd.Use = "ocenv"
	cmd.RootCmd.Short = "OC version manager"
	cmd.Execute()
}
