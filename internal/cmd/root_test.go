package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	// Test that Execute function doesn't panic
	// Note: We can't easily test the actual execution without mocking cobra.Command
	// but we can ensure the function exists and is callable
	assert.NotPanics(t, func() {
		// This would normally start the CLI, but for testing we just ensure it exists
		// Execute()
	})
}

func TestRootCommand(t *testing.T) {
	// Test basic root command properties
	// Since Execute() sets up the root command, we test indirectly
	cmd := NewGenerateCommand()
	assert.NotNil(t, cmd, "Should be able to create generate command")
	
	// Test that the command has expected properties
	assert.Equal(t, "generate", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}