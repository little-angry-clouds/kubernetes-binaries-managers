package main

import (
	"github.com/nixknight/binaries-managers/internal/cmd"
)

var binaryDownloadURL string = "https://storage.googleapis.com/kubernetes-release/release/v%s/bin/%s/%s/kubectl" // nolint:lll
var versionsAPI string = "https://api.github.com/repos/kubernetes/kubernetes/releases?per_page=100&page="        // nolint:lll

func main() {
	cmd.BinaryDownloadURL = binaryDownloadURL
	cmd.VersionsAPI = versionsAPI
	cmd.BinaryToInstall = "kubectl"
	cmd.RootCmd.Use = "kbenv"
	cmd.RootCmd.Short = "Kubectl version manager"
	cmd.Execute()
}
