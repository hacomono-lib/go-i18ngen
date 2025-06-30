// Package config handles configuration loading and management for the i18n code generator.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds configuration for i18ngen
type Config struct {
	Locales            []string `yaml:"locales"`
	Compound           bool     `yaml:"compound"`
	MessagesGlob       string   `yaml:"messages"`
	PlaceholdersGlob   string   `yaml:"placeholders"`
	OutputDir          string   `yaml:"output_dir"`
	OutputPackage      string   `yaml:"output_package"`
	PluralPlaceholders []string `yaml:"plural_placeholders"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Return empty config if file doesn't exist
		return &Config{}, nil
	}

	// Start with default configuration for existing files
	config := &Config{
		Locales:            []string{"en", "ja"},
		Compound:           true,
		MessagesGlob:       "./messages/*.yaml",
		PlaceholdersGlob:   "./placeholders/*.yaml",
		OutputDir:          "./",
		OutputPackage:      "i18n",
		PluralPlaceholders: getDefaultPluralPlaceholders(),
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}

	// Resolve relative paths based on config file directory
	configDir := filepath.Dir(path)
	if config.MessagesGlob != "" && !filepath.IsAbs(config.MessagesGlob) {
		config.MessagesGlob = filepath.Join(configDir, config.MessagesGlob)
	}
	if config.PlaceholdersGlob != "" && !filepath.IsAbs(config.PlaceholdersGlob) {
		config.PlaceholdersGlob = filepath.Join(configDir, config.PlaceholdersGlob)
	}
	if config.OutputDir != "" && !filepath.IsAbs(config.OutputDir) {
		config.OutputDir = filepath.Join(configDir, config.OutputDir)
	}

	return config, nil
}

// getDefaultPluralPlaceholders returns the default list of placeholder names
// that should be treated as plural count fields
func getDefaultPluralPlaceholders() []string {
	return []string{
		"Count", // Standard pluralization placeholder: {{.Count}}
	}
}

// GetPluralPlaceholders returns the configured plural placeholder names
// with case variations (Count, count, PluralCount, etc.)
func (c *Config) GetPluralPlaceholders() []string {
	if len(c.PluralPlaceholders) == 0 {
		c.PluralPlaceholders = getDefaultPluralPlaceholders()
	}

	var result []string
	for _, placeholder := range c.PluralPlaceholders {
		// Add the original name
		result = append(result, placeholder)

		// Add lowercase version
		if placeholder != strings.ToLower(placeholder) {
			result = append(result, strings.ToLower(placeholder))
		}
	}

	return result
}

// IsPluralPlaceholder checks if a placeholder name is configured as a plural placeholder
// (case-insensitive comparison)
func (c *Config) IsPluralPlaceholder(name string) bool {
	plurals := c.GetPluralPlaceholders()

	for _, placeholder := range plurals {
		if strings.EqualFold(placeholder, name) {
			return true
		}
	}

	return false
}
