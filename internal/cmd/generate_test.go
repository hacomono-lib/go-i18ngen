package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeConfig(t *testing.T) {
	t.Run("command line arguments override config.yaml values", func(t *testing.T) {
		// config.yaml settings
		cfg := &config.Config{
			Locales:          []string{"ja"},
			Compound:         false,
			MessagesGlob:     "/config/messages/*.json",
			PlaceholdersGlob: "/config/placeholders/*.yaml",
			OutputDir:        "/config/output",
			OutputPackage:    "config_pkg",
		}

		// command line flags
		flags := &Flags{
			Locales:          []string{"ja", "en"},
			Compound:         true,
			MessagesGlob:     "/cmd/messages/*.json",
			PlaceholdersGlob: "/cmd/placeholders/*.yaml",
			OutputDir:        "/cmd/output",
			OutputPackage:    "cmd_pkg",
		}

		merged := MergeConfig(cfg, flags)

		// verify that command line argument values take precedence
		assert.Equal(t, []string{"ja", "en"}, merged.Locales)
		assert.True(t, merged.Compound)
		assert.Equal(t, "/cmd/messages/*.json", merged.MessagesGlob)
		assert.Equal(t, "/cmd/placeholders/*.yaml", merged.PlaceholdersGlob)
		assert.Equal(t, "/cmd/output", merged.OutputDir)
		assert.Equal(t, "cmd_pkg", merged.OutputPackage)
	})

	t.Run("uses config.yaml values when command line arguments are empty", func(t *testing.T) {
		// config.yaml settings
		cfg := &config.Config{
			Locales:          []string{"ja"},
			Compound:         true,
			MessagesGlob:     "/config/messages/*.json",
			PlaceholdersGlob: "/config/placeholders/*.yaml",
			OutputDir:        "/config/output",
			OutputPackage:    "config_pkg",
		}

		// empty command line flags
		flags := &Flags{}

		merged := MergeConfig(cfg, flags)

		// verify that config.yaml values are used as-is
		assert.Equal(t, []string{"ja"}, merged.Locales)
		assert.True(t, merged.Compound)
		assert.Equal(t, "/config/messages/*.json", merged.MessagesGlob)
		assert.Equal(t, "/config/placeholders/*.yaml", merged.PlaceholdersGlob)
		assert.Equal(t, "/config/output", merged.OutputDir)
		assert.Equal(t, "config_pkg", merged.OutputPackage)
	})

	t.Run("partial command line argument override", func(t *testing.T) {
		// config.yaml settings
		cfg := &config.Config{
			Locales:          []string{"ja"},
			Compound:         false,
			MessagesGlob:     "/config/messages/*.json",
			PlaceholdersGlob: "/config/placeholders/*.yaml",
			OutputDir:        "/config/output",
			OutputPackage:    "config_pkg",
		}

		// specify only some command line arguments
		flags := &Flags{
			MessagesGlob: "/cmd/messages/*.json",
			OutputDir:    "/cmd/output",
		}

		merged := MergeConfig(cfg, flags)

		// only specified command line arguments are overridden, others use config.yaml values
		assert.Equal(t, []string{"ja"}, merged.Locales)                         // config.yaml value
		assert.False(t, merged.Compound)                                        // config.yaml value
		assert.Equal(t, "/cmd/messages/*.json", merged.MessagesGlob)            // overridden by command line
		assert.Equal(t, "/config/placeholders/*.yaml", merged.PlaceholdersGlob) // config.yaml value
		assert.Equal(t, "/cmd/output", merged.OutputDir)                        // overridden by command line
		assert.Equal(t, "config_pkg", merged.OutputPackage)                     // config.yaml value
	})
}

func TestPathResolutionBehavior(t *testing.T) {
	// create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_path_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// create subdirectory
	configDir := filepath.Join(tempDir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0750))

	t.Run("paths in config.yaml are relative to config file directory", func(t *testing.T) {
		// create config.yaml file
		configPath := filepath.Join(configDir, "test_config.yaml")
		configContent := `messages: "messages/*.json"
placeholders: "placeholders/*.yaml"
output_dir: "output"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// load config.yaml
		cfg, err := config.LoadConfig(configPath)
		require.NoError(t, err)

		// verify that paths are resolved relative to the config file directory
		expectedMessagesGlob := filepath.Join(configDir, "messages/*.json")
		expectedPlaceholdersGlob := filepath.Join(configDir, "placeholders/*.yaml")
		expectedOutputDir := filepath.Join(configDir, "output")

		assert.Equal(t, expectedMessagesGlob, cfg.MessagesGlob)
		assert.Equal(t, expectedPlaceholdersGlob, cfg.PlaceholdersGlob)
		assert.Equal(t, expectedOutputDir, cfg.OutputDir)
	})

	t.Run("command line paths are used as-is (relative to execution directory)", func(t *testing.T) {
		// create empty config.yaml
		configPath := filepath.Join(configDir, "empty_config.yaml")
		configContent := `locales: ["ja"]`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// load config.yaml
		cfg, err := config.LoadConfig(configPath)
		require.NoError(t, err)

		// command line flags (relative to execution directory)
		flags := &Flags{
			MessagesGlob:     "cmd_messages/*.json",     // from execution directory
			PlaceholdersGlob: "cmd_placeholders/*.yaml", // from execution directory
			OutputDir:        "cmd_output",              // from execution directory
		}

		merged := MergeConfig(cfg, flags)

		// command line paths are used as-is (no path resolution)
		assert.Equal(t, "cmd_messages/*.json", merged.MessagesGlob)
		assert.Equal(t, "cmd_placeholders/*.yaml", merged.PlaceholdersGlob)
		assert.Equal(t, "cmd_output", merged.OutputDir)
	})
}

func TestNewGenerateCommand(t *testing.T) {
	cmd := NewGenerateCommand()

	assert.Equal(t, "generate", cmd.Use)
	assert.Contains(t, cmd.Short, "Generate i18n message and placeholder code")
	assert.NotNil(t, cmd.RunE)

	// Check that flags are properly defined
	assert.NotNil(t, cmd.Flags().Lookup("config"))
	assert.NotNil(t, cmd.Flags().Lookup("locales"))
	assert.NotNil(t, cmd.Flags().Lookup("compound"))
	assert.NotNil(t, cmd.Flags().Lookup("messages"))
	assert.NotNil(t, cmd.Flags().Lookup("placeholders"))
	assert.NotNil(t, cmd.Flags().Lookup("output"))
	assert.NotNil(t, cmd.Flags().Lookup("package"))
}

func TestGenerateCommandExecution(t *testing.T) {
	tempDir := t.TempDir()

	// Create test config file
	configContent := `
compound: true
locales: [ja, en]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "i18n"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create directories
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	err = os.MkdirAll(messagesDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(placeholdersDir, 0755)
	require.NoError(t, err)

	// Create test message file
	messageContent := `
TestMessage:
  ja: "テストメッセージ: {{.field}}"
  en: "Test message: {{.field}}"
`
	messagePath := filepath.Join(messagesDir, "test.yaml")
	err = os.WriteFile(messagePath, []byte(messageContent), 0644)
	require.NoError(t, err)

	// Create test placeholder file
	placeholderContent := `
example:
  ja: "例"
  en: "example"
`
	placeholderPath := filepath.Join(placeholdersDir, "field.yaml")
	err = os.WriteFile(placeholderPath, []byte(placeholderContent), 0644)
	require.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create and execute command
	cmd := NewGenerateCommand()
	cmd.SetArgs([]string{"--config", configPath})

	err = cmd.Execute()
	assert.NoError(t, err)

	// Verify output file was created
	outputPath := filepath.Join(tempDir, "i18n.gen.go")
	_, err = os.Stat(outputPath)
	assert.NoError(t, err, "Generated file should exist")

	// Verify content contains expected elements
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "package i18n")
	assert.Contains(t, contentStr, "type TestMessage struct")
	assert.Contains(t, contentStr, "type FieldText struct")
}

func TestGenerateCommandWithFlags(t *testing.T) {
	tempDir := t.TempDir()

	// Create directories
	messagesDir := filepath.Join(tempDir, "msgs")
	placeholdersDir := filepath.Join(tempDir, "phs")
	err := os.MkdirAll(messagesDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(placeholdersDir, 0755)
	require.NoError(t, err)

	// Create test files
	messageContent := `
SimpleMessage:
  en: "Hello {{.name}}"
  fr: "Bonjour {{.name}}"
`
	messagePath := filepath.Join(messagesDir, "simple.yaml")
	err = os.WriteFile(messagePath, []byte(messageContent), 0644)
	require.NoError(t, err)

	placeholderContent := `
john:
  en: "John"
  fr: "Jean"
`
	placeholderPath := filepath.Join(placeholdersDir, "name.yaml")
	err = os.WriteFile(placeholderPath, []byte(placeholderContent), 0644)
	require.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create and execute command with flags
	cmd := NewGenerateCommand()
	cmd.SetArgs([]string{
		"--locales", "en,fr",
		"--compound", "true",
		"--messages", "msgs/*.yaml",
		"--placeholders", "phs/*.yaml",
		"--output", ".",
		"--package", "testpkg",
	})

	err = cmd.Execute()
	assert.NoError(t, err)

	// Verify output file was created with correct package name
	outputPath := filepath.Join(tempDir, "i18n.gen.go")
	_, err = os.Stat(outputPath)
	assert.NoError(t, err, "Generated file should exist")

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "package testpkg")
}

func TestGenerateCommandError(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create command with invalid config
	cmd := NewGenerateCommand()
	cmd.SetArgs([]string{"--config", "nonexistent.yaml"})

	err = cmd.Execute()
	assert.Error(t, err, "Should fail with nonexistent config file")
}
