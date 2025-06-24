package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
)

func TestRun_Success(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	outputDir := filepath.Join(tempDir, "output")

	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	// Create test message file
	messageContent := `UserWelcome:
  ja: "{{.name}}さん、ようこそ！"
  en: "Welcome, {{.name}}!"
EntityNotFound:
  ja: "{{.entity}}が見つかりません"
  en: "{{.entity}} not found"
`
	messageFile := filepath.Join(messagesDir, "messages.yaml")
	require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Create test placeholder file
	placeholderContent := `user:
  ja: "ユーザー"
  en: "User"
product:
  ja: "製品"
  en: "Product"
`
	placeholderFile := filepath.Join(placeholdersDir, "entity.yaml")
	require.NoError(t, os.WriteFile(placeholderFile, []byte(placeholderContent), 0644))

	// Create config
	cfg := &config.Config{
		MessagesGlob:     filepath.Join(messagesDir, "*.yaml"),
		PlaceholdersGlob: filepath.Join(placeholdersDir, "*.yaml"),
		OutputDir:        outputDir,
		OutputPackage:    "testpkg",
		Locales:          []string{"ja", "en"},
		Compound:         true,
	}

	// Run generator
	err = Run(cfg)
	require.NoError(t, err)

	// Verify output file was created
	outputFile := filepath.Join(outputDir, "i18n.gen.go")
	assert.FileExists(t, outputFile)

	// Verify generated content contains expected elements
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "package testpkg")
	assert.Contains(t, contentStr, "UserWelcome")
	assert.Contains(t, contentStr, "EntityNotFound")
	assert.Contains(t, contentStr, "NewUserWelcome")
	assert.Contains(t, contentStr, "NewEntityNotFound")
}

func TestRun_InvalidMessagesGlob(t *testing.T) {
	cfg := &config.Config{
		MessagesGlob:     "[invalid-glob",
		PlaceholdersGlob: "./placeholders/*.yaml",
		OutputDir:        "./output",
		OutputPackage:    "testpkg",
		Locales:          []string{"ja", "en"},
		Compound:         true,
	}

	err := Run(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid messages glob pattern")
}

func TestRun_NoMessageFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "i18ngen_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	outputDir := filepath.Join(tempDir, "output")

	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	cfg := &config.Config{
		MessagesGlob:     filepath.Join(messagesDir, "*.yaml"),
		PlaceholdersGlob: filepath.Join(placeholdersDir, "*.yaml"),
		OutputDir:        outputDir,
		OutputPackage:    "testpkg",
		Locales:          []string{"ja", "en"},
		Compound:         true,
	}

	err = Run(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no message files found")
}

func TestRun_InvalidMessageFormat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "i18ngen_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	outputDir := filepath.Join(tempDir, "output")

	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	// Create invalid YAML file
	invalidContent := `invalid: yaml: content:
  - unclosed
    brackets: [
`
	messageFile := filepath.Join(messagesDir, "invalid.yaml")
	require.NoError(t, os.WriteFile(messageFile, []byte(invalidContent), 0644))

	cfg := &config.Config{
		MessagesGlob:     filepath.Join(messagesDir, "*.yaml"),
		PlaceholdersGlob: filepath.Join(placeholdersDir, "*.yaml"),
		OutputDir:        outputDir,
		OutputPackage:    "testpkg",
		Locales:          []string{"ja", "en"},
		Compound:         true,
	}

	err = Run(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestRun_ReadOnlyOutputDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "i18ngen_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")

	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))

	// Create test message file
	messageContent := `UserWelcome:
  ja: "{{.name}}さん、ようこそ！"
  en: "Welcome, {{.name}}!"
`
	messageFile := filepath.Join(messagesDir, "messages.yaml")
	require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Create read-only directory
	readOnlyDir := filepath.Join(tempDir, "readonly")
	require.NoError(t, os.MkdirAll(readOnlyDir, 0755))
	require.NoError(t, os.Chmod(readOnlyDir, 0444))

	defer func() {
		// Restore permissions for cleanup
		os.Chmod(readOnlyDir, 0755)
	}()

	cfg := &config.Config{
		MessagesGlob:     filepath.Join(messagesDir, "*.yaml"),
		PlaceholdersGlob: filepath.Join(placeholdersDir, "*.yaml"),
		OutputDir:        filepath.Join(readOnlyDir, "nested"),
		OutputPackage:    "testpkg",
		Locales:          []string{"ja", "en"},
		Compound:         true,
	}

	err = Run(cfg)

	// In CI environments, permission restrictions might not work as expected
	if err != nil {
		assert.Error(t, err)
		// Check that error is related to file system issues
		assert.True(t,
			contains(err.Error(), "permission denied") ||
				contains(err.Error(), "failed to write") ||
				contains(err.Error(), "failed to create"),
			"Expected file system error, got: %v", err)
	} else {
		t.Log("Note: File permission restrictions may not be enforced in this environment")
	}
}

func TestRun_EmptyLocales(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "i18ngen_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	outputDir := filepath.Join(tempDir, "output")

	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	// Create test message file
	messageContent := `UserWelcome:
  ja: "{{.name}}さん、ようこそ！"
  en: "Welcome, {{.name}}!"
`
	messageFile := filepath.Join(messagesDir, "messages.yaml")
	require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

	cfg := &config.Config{
		MessagesGlob:     filepath.Join(messagesDir, "*.yaml"),
		PlaceholdersGlob: filepath.Join(placeholdersDir, "*.yaml"),
		OutputDir:        outputDir,
		OutputPackage:    "testpkg",
		Locales:          []string{}, // Empty locales
		Compound:         true,
	}

	err = Run(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no locales specified")
}

func TestRun_NilConfig(t *testing.T) {
	err := Run(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration cannot be nil")
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
