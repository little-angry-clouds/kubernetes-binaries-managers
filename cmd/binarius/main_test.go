package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportedToolsIntegration(t *testing.T) {
	// Test that all supported tools have properly configured subcommands
	allTools := getAllTools()
	assert.Len(t, allTools, 3, "Should support exactly 3 tools")
	assert.Contains(t, allTools, "kubectl")
	assert.Contains(t, allTools, "helm")
	assert.Contains(t, allTools, "oc")
}

func TestToolConfigurationIntegration(t *testing.T) {
	// Test that each tool has valid configuration
	for _, toolName := range getAllTools() {
		toolConfig, exists := getTool(toolName)
		assert.True(t, exists, "Tool %s should exist in configuration", toolName)
		assert.NotEmpty(t, toolConfig.Name, "Tool %s should have a name", toolName)
		assert.NotEmpty(t, toolConfig.DownloadURL, "Tool %s should have a download URL", toolName)
		assert.NotEmpty(t, toolConfig.VersionsAPI, "Tool %s should have a versions API", toolName)
		assert.NotEmpty(t, toolConfig.Description, "Tool %s should have a description", toolName)
		assert.NotEmpty(t, toolConfig.CompressedType, "Tool %s should have a compression type", toolName)
		assert.Contains(t, []string{"binary", "tar.gz", "zip"}, toolConfig.CompressedType, "Tool %s should have valid compression type", toolName)
	}
}
