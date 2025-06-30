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
			name: "default plural placeholders - exact match",
			config: &Config{
				PluralPlaceholders: []string{"Count", "Number", "Total"},
			},
			placeholder: "Count",
			expected:    true,
		},
		{
			name: "default plural placeholders - case insensitive",
			config: &Config{
				PluralPlaceholders: []string{"Count", "Number", "Total"},
			},
			placeholder: "count",
			expected:    true,
		},
		{
			name: "default plural placeholders - uppercase",
			config: &Config{
				PluralPlaceholders: []string{"Count", "Number", "Total"},
			},
			placeholder: "COUNT",
			expected:    true,
		},
		{
			name: "non-plural placeholder",
			config: &Config{
				PluralPlaceholders: []string{"Count", "Number", "Total"},
			},
			placeholder: "Name",
			expected:    false,
		},
		{
			name: "empty config uses defaults",
			config: &Config{
				PluralPlaceholders: []string{}, // Empty, should use defaults
			},
			placeholder: "Count",
			expected:    true,
		},
		{
			name: "nil config uses defaults",
			config: &Config{
				PluralPlaceholders: nil, // Nil, should use defaults
			},
			placeholder: "Count",
			expected:    true,
		},
		{
			name: "custom plural placeholders",
			config: &Config{
				PluralPlaceholders: []string{"CustomCount", "Items"},
			},
			placeholder: "CustomCount",
			expected:    true,
		},
		{
			name: "custom plural placeholders - case insensitive",
			config: &Config{
				PluralPlaceholders: []string{"CustomCount", "Items"},
			},
			placeholder: "customcount",
			expected:    true,
		},
		{
			name: "default not in custom list",
			config: &Config{
				PluralPlaceholders: []string{"CustomCount", "Items"},
			},
			placeholder: "Count", // Default, but not in custom list
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

func (s *ConfigTestSuite) TestGetPluralPlaceholders() {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name: "default plural placeholders",
			config: &Config{
				PluralPlaceholders: []string{},
			},
			expected: []string{
				"Count", "count",
			},
		},
		{
			name: "custom plural placeholders",
			config: &Config{
				PluralPlaceholders: []string{"Items", "Records"},
			},
			expected: []string{
				"Items", "items",
				"Records", "records",
			},
		},
		{
			name: "mixed case custom placeholders",
			config: &Config{
				PluralPlaceholders: []string{"Count", "itemCount"},
			},
			expected: []string{
				"Count", "count",
				"itemCount",
			},
		},
		{
			name: "nil config uses defaults",
			config: &Config{
				PluralPlaceholders: []string{},
			},
			expected: []string{
				"Count", "count",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := tt.config.GetPluralPlaceholders()

			// Check that all expected placeholders are present
			for _, expected := range tt.expected {
				s.Contains(result, expected, "Expected placeholder %q not found", expected)
			}

			// For default config, check that we have the right number
			if len(tt.config.PluralPlaceholders) == 0 {
				// Should have Count + count (case variations)
				s.GreaterOrEqual(len(result), len(tt.expected), "Should have at least expected number of placeholders")
			}
		})
	}
}

func (s *ConfigTestSuite) TestGetDefaultPluralPlaceholders() {
	defaults := getDefaultPluralPlaceholders()

	expected := []string{"Count"}
	s.Equal(expected, defaults)
}

func (s *ConfigTestSuite) TestLoadConfigWithPluralPlaceholders() {
	// Create a temporary config file
	configPath := filepath.Join(s.tempDir, "config.yaml")

	configContent := `
locales: ["en", "ja"]
plural_placeholders: ["Count", "Number", "CustomTotal"]
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	// Load the config
	config, err := LoadConfig(configPath)
	s.Require().NoError(err)

	// Verify plural placeholders are loaded correctly
	s.Equal([]string{"Count", "Number", "CustomTotal"}, config.PluralPlaceholders)
	s.True(config.IsPluralPlaceholder("Count"))
	s.True(config.IsPluralPlaceholder("count"))
	s.True(config.IsPluralPlaceholder("CustomTotal"))
	s.True(config.IsPluralPlaceholder("customtotal"))
	s.False(config.IsPluralPlaceholder("Amount")) // Not in custom list
}

func (s *ConfigTestSuite) TestLoadConfigWithoutPluralPlaceholders() {
	// Create a temporary config file without plural_placeholders
	configPath := filepath.Join(s.tempDir, "config.yaml")

	configContent := `
locales: ["en", "ja"]
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	// Load the config
	config, err := LoadConfig(configPath)
	s.Require().NoError(err)

	// Verify default plural placeholders are used
	s.Equal(getDefaultPluralPlaceholders(), config.PluralPlaceholders)
	s.True(config.IsPluralPlaceholder("Count"))
	s.True(config.IsPluralPlaceholder("count"))
}

func (s *ConfigTestSuite) TestLoadConfigWithEmptyPluralPlaceholders() {
	configPath := filepath.Join(s.tempDir, "config_empty.yaml")
	configContent := `
locales: ["en", "ja"]
plural_placeholders: []
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	config, err := LoadConfig(configPath)
	s.Require().NoError(err)

	// Empty list should trigger default behavior
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
		PluralPlaceholders: []string{"Count", "Number"},
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
