// Package config handles configuration loading and management for the i18n code generator.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultPluralPlaceholder is the default plural placeholder name
	DefaultPluralPlaceholder = "Count"
)

// Config holds configuration for i18ngen
type Config struct {
	Locales           []string `yaml:"locales"`
	Compound          bool     `yaml:"compound"`
	MessagesGlob      string   `yaml:"messages"`
	PlaceholdersGlob  string   `yaml:"placeholders"`
	OutputDir         string   `yaml:"output_dir"`
	OutputPackage     string   `yaml:"output_package"`
	PluralPlaceholder string   `yaml:"plural_placeholder"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path) // #nosec G304 - Reading configuration file is intentional
	if err != nil {
		// Return empty config if file doesn't exist
		return &Config{}, nil
	}

	// Start with default configuration for existing files
	config := &Config{
		Locales:           []string{"en", "ja"},
		Compound:          true,
		MessagesGlob:      "./messages/*.yaml",
		PlaceholdersGlob:  "./placeholders/*.yaml",
		OutputDir:         "./",
		OutputPackage:     "i18n",
		PluralPlaceholder: DefaultPluralPlaceholder,
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

// GetPluralPlaceholder returns the configured plural placeholder name
func (c *Config) GetPluralPlaceholder() string {
	if c.PluralPlaceholder == "" {
		return DefaultPluralPlaceholder // Default value
	}
	return c.PluralPlaceholder
}

// IsPluralPlaceholder checks if a placeholder name is the configured plural placeholder (case-insensitive)
func (c *Config) IsPluralPlaceholder(name string) bool {
	return strings.EqualFold(name, c.GetPluralPlaceholder())
}
