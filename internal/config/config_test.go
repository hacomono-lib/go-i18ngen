package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	tempDir string
}

func (s *ConfigTestSuite) SetupSuite() {
	s.tempDir = s.T().TempDir()
}

func (s *ConfigTestSuite) TestIsPluralPlaceholder() {
	tests := []struct {
		name        string
		config      *Config
		placeholder string
		expected    bool
	}{
		{
			name: "default plural placeholder - exact match",
			config: &Config{
				PluralPlaceholder: "", // Should use default "Count"
			},
			placeholder: "Count",
			expected:    true,
		},
		{
			name: "default plural placeholder - case insensitive",
			config: &Config{
				PluralPlaceholder: "",
			},
			placeholder: "count",
			expected:    true,
		},
		{
			name: "default plural placeholder - uppercase",
			config: &Config{
				PluralPlaceholder: "",
			},
			placeholder: "COUNT",
			expected:    true,
		},
		{
			name: "non-plural placeholder with default",
			config: &Config{
				PluralPlaceholder: "",
			},
			placeholder: "Name",
			expected:    false,
		},
		{
			name: "custom plural placeholder",
			config: &Config{
				PluralPlaceholder: "CustomCount",
			},
			placeholder: "CustomCount",
			expected:    true,
		},
		{
			name: "custom plural placeholder - case insensitive",
			config: &Config{
				PluralPlaceholder: "CustomCount",
			},
			placeholder: "customcount",
			expected:    true,
		},
		{
			name: "default not matching custom",
			config: &Config{
				PluralPlaceholder: "CustomCount",
			},
			placeholder: "Count", // Default, but not the custom one
			expected:    false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := tt.config.IsPluralPlaceholder(tt.placeholder)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ConfigTestSuite) TestGetPluralPlaceholder() {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "default plural placeholder",
			config: &Config{
				PluralPlaceholder: "",
			},
			expected: "Count",
		},
		{
			name: "custom plural placeholder",
			config: &Config{
				PluralPlaceholder: "CustomCount",
			},
			expected: "CustomCount",
		},
		{
			name: "empty string placeholder",
			config: &Config{
				PluralPlaceholder: "",
			},
			expected: "Count", // Should return default when empty
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := tt.config.GetPluralPlaceholder()
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ConfigTestSuite) TestLoadConfigWithPluralPlaceholder() {
	// Create a temporary config file
	configPath := filepath.Join(s.tempDir, "config.yaml")

	configContent := `
locales: ["en", "ja"]
plural_placeholder: "CustomTotal"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	// Load the config
	config, err := LoadConfig(configPath)
	s.Require().NoError(err)

	// Verify plural placeholder is loaded correctly
	s.Equal("CustomTotal", config.GetPluralPlaceholder())
	s.True(config.IsPluralPlaceholder("CustomTotal"))
	s.True(config.IsPluralPlaceholder("customtotal"))
	s.False(config.IsPluralPlaceholder("Count")) // Not the custom one
}

func (s *ConfigTestSuite) TestLoadConfigWithoutPluralPlaceholder() {
	// Create a temporary config file without plural_placeholder
	configPath := filepath.Join(s.tempDir, "config.yaml")

	configContent := `
locales: ["en", "ja"]
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	// Load the config
	config, err := LoadConfig(configPath)
	s.Require().NoError(err)

	// Verify default plural placeholder is used
	s.Equal("Count", config.GetPluralPlaceholder())
	s.True(config.IsPluralPlaceholder("Count"))
	s.True(config.IsPluralPlaceholder("count"))
}

func (s *ConfigTestSuite) TestLoadConfigFileNotExists() {
	nonExistentPath := filepath.Join(s.tempDir, "nonexistent.yaml")

	config, err := LoadConfig(nonExistentPath)
	s.Require().NoError(err)

	// Should return empty config when file doesn't exist
	s.Empty(config.Locales)
}

func (s *ConfigTestSuite) TestLoadConfigInvalidYAML() {
	configPath := filepath.Join(s.tempDir, "invalid.yaml")
	invalidContent := `
invalid: yaml: content:
  - missing
    proper: structure
`

	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	s.Require().NoError(err)

	_, err = LoadConfig(configPath)
	s.Error(err)
	s.Contains(err.Error(), "failed to parse config file")
}

func (s *ConfigTestSuite) TestConfigPathResolution() {
	// Create a subdirectory
	subDir := filepath.Join(s.tempDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	s.Require().NoError(err)

	configPath := filepath.Join(subDir, "config.yaml")
	configContent := `
locales: ["en", "ja"]
messages: "../messages/*.yaml"
placeholders: "../placeholders/*.yaml"
output_dir: "../output"
`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	config, err := LoadConfig(configPath)
	s.Require().NoError(err)

	// Paths should be resolved relative to config file directory
	s.Equal(filepath.Join(s.tempDir, "messages", "*.yaml"), config.MessagesGlob)
	s.Equal(filepath.Join(s.tempDir, "placeholders", "*.yaml"), config.PlaceholdersGlob)
	s.Equal(filepath.Join(s.tempDir, "output"), config.OutputDir)
}

func (s *ConfigTestSuite) TestConfigWithAbsolutePaths() {
	configPath := filepath.Join(s.tempDir, "config_abs.yaml")
	absPath := "/absolute/path/messages/*.yaml"
	configContent := `
locales: ["en", "ja"]
messages: "` + absPath + `"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	config, err := LoadConfig(configPath)
	s.Require().NoError(err)

	// Absolute paths should remain unchanged
	s.Equal(absPath, config.MessagesGlob)
}

func (s *ConfigTestSuite) TestPluralPlaceholderEdgeCases() {
	config := &Config{
		PluralPlaceholder: "Count",
	}

	// Test empty string
	s.False(config.IsPluralPlaceholder(""))

	// Test whitespace
	s.False(config.IsPluralPlaceholder(" "))
	s.False(config.IsPluralPlaceholder("Count "))
	s.False(config.IsPluralPlaceholder(" Count"))

	// Test special characters
	s.False(config.IsPluralPlaceholder("Count!"))
	s.False(config.IsPluralPlaceholder("Count-1"))

	// Test substring matches (should not match)
	s.False(config.IsPluralPlaceholder("Coun"))
	s.False(config.IsPluralPlaceholder("ount"))
	s.False(config.IsPluralPlaceholder("MyCount"))
	s.False(config.IsPluralPlaceholder("CountValue"))
}

// Run the test suite
func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
