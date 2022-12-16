package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersionString(t *testing.T) {
	var flagtests = []struct {
		testName string
		input    string
		output   string
	}{
		// vanilla Kubernetes (ie: kubeadm)
		{"stable without suffix", "v1.20.0", "1.20.0"},
		{"alpha without suffix", "v1.2.0-alpha.6", "1.2.0-alpha.6"},
		{"beta without suffix", "v1.26.0-beta.0", "1.26.0-beta.0"},
		{"rc without suffix", "v1.26.0-rc.1", "1.26.0-rc.1"},

		// AWS EKS version strings
		{"stable with EKS suffix", "v1.21.14-eks-abcdefg", "1.21.14"},
		{"alpha with EKS suffix", "v1.21.0-alpha.1-eks-abcdefg", "1.21.0-alpha.1"},
		{"beta with EKS suffix", "v1.21.0-beta.1-eks-abcdefg", "1.21.0-beta.1"},
		{"rc with EKS suffix", "v1.21.0-rc.1-eks-abcdefg", "1.21.0-rc.1"},

		// GCP GKE version strings (also Okteto)
		{"stable with GKE suffix", "v1.23.11-gke.300", "1.23.11"},
		{"alpha with GKE suffix", "v1.23.0-alpha.1-gke.300", "1.23.0-alpha.1"},
		{"beta with GKE suffix", "v1.23.0-beta.1-gke.300", "1.23.0-beta.1"},
		{"rc with GKE suffix", "v1.23.0-rc.1-gke.300", "1.23.0-rc.1"},
	}

	for _, tt := range flagtests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			parsedVersion := parseVersionString(tt.input)
			assert.Equal(t, tt.output, parsedVersion)
		})
	}
}
