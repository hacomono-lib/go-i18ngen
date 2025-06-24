package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// create subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	// create config file in subdirectory
	configPath := filepath.Join(subDir, "test_config.yaml")
	configContent := `locales:
  - ja
  - en
compound: true
messages: "messages/*.json"
placeholders: "placeholders/*.yaml"
output_dir: "output"
output_package: "i18n"
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	t.Run("relative path resolution", func(t *testing.T) {
		config, err := LoadConfig(configPath)
		require.NoError(t, err)

		// verify that relative paths are resolved based on the config file directory
		expectedMessagesGlob := filepath.Join(subDir, "messages/*.json")
		expectedPlaceholdersGlob := filepath.Join(subDir, "placeholders/*.yaml")
		expectedOutputDir := filepath.Join(subDir, "output")

		assert.Equal(t, []string{"ja", "en"}, config.Locales)
		assert.True(t, config.Compound)
		assert.Equal(t, expectedMessagesGlob, config.MessagesGlob)
		assert.Equal(t, expectedPlaceholdersGlob, config.PlaceholdersGlob)
		assert.Equal(t, expectedOutputDir, config.OutputDir)
		assert.Equal(t, "i18n", config.OutputPackage)
	})

	t.Run("absolute paths remain unchanged", func(t *testing.T) {
		// create config file with absolute paths
		absoluteConfigPath := filepath.Join(subDir, "absolute_config.yaml")
		absoluteConfigContent := `locales:
  - ja
compound: false
messages: "/absolute/path/messages/*.json"
placeholders: "/absolute/path/placeholders/*.yaml"
output_dir: "/absolute/path/output"
output_package: "i18n"
`
		require.NoError(t, os.WriteFile(absoluteConfigPath, []byte(absoluteConfigContent), 0644))

		config, err := LoadConfig(absoluteConfigPath)
		require.NoError(t, err)

		// verify that absolute paths are preserved as-is
		assert.Equal(t, []string{"ja"}, config.Locales)
		assert.False(t, config.Compound)
		assert.Equal(t, "/absolute/path/messages/*.json", config.MessagesGlob)
		assert.Equal(t, "/absolute/path/placeholders/*.yaml", config.PlaceholdersGlob)
		assert.Equal(t, "/absolute/path/output", config.OutputDir)
		assert.Equal(t, "i18n", config.OutputPackage)
	})

	t.Run("non-existent file returns default config", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "non_existent.yaml")
		config, err := LoadConfig(nonExistentPath)
		require.NoError(t, err)

		// verify that default (empty) config is returned
		assert.Empty(t, config.Locales)
		assert.False(t, config.Compound)
		assert.Empty(t, config.MessagesGlob)
		assert.Empty(t, config.PlaceholdersGlob)
		assert.Empty(t, config.OutputDir)
		assert.Empty(t, config.OutputPackage)
	})

	t.Run("invalid YAML file", func(t *testing.T) {
		invalidConfigPath := filepath.Join(subDir, "invalid_config.yaml")
		invalidContent := `invalid: yaml: content:
  - unclosed
    brackets: [
`
		require.NoError(t, os.WriteFile(invalidConfigPath, []byte(invalidContent), 0644))

		config, err := LoadConfig(invalidConfigPath)
		assert.Error(t, err)
		assert.Nil(t, config)
	})

	t.Run("empty string paths are not resolved", func(t *testing.T) {
		emptyConfigPath := filepath.Join(subDir, "empty_config.yaml")
		emptyConfigContent := `locales:
  - ja
compound: true
messages: ""
placeholders: ""
output_dir: ""
output_package: "i18n"
`
		require.NoError(t, os.WriteFile(emptyConfigPath, []byte(emptyConfigContent), 0644))

		config, err := LoadConfig(emptyConfigPath)
		require.NoError(t, err)

		// empty strings are preserved as-is
		assert.Equal(t, []string{"ja"}, config.Locales)
		assert.True(t, config.Compound)
		assert.Equal(t, "", config.MessagesGlob)
		assert.Equal(t, "", config.PlaceholdersGlob)
		assert.Equal(t, "", config.OutputDir)
		assert.Equal(t, "i18n", config.OutputPackage)
	})
}
