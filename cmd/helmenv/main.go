package main

import (
	"github.com/nixknight/binaries-managers/internal/cmd"
)

var binaryDownloadURL string = "https://get.helm.sh/helm-v%s-%s-%s"                           // nolint:lll
var versionsAPI string = "https://api.github.com/repos/helm/helm/releases?per_page=100&page=" // nolint:lll

func main() {
	cmd.BinaryDownloadURL = binaryDownloadURL
	cmd.VersionsAPI = versionsAPI
	cmd.BinaryToInstall = "helm"
	cmd.RootCmd.Use = "helmenv"
	cmd.RootCmd.Short = "Helm version manager"
	cmd.Execute()
}
