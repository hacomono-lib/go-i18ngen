// Package generator implements the main code generation logic for i18n messages and placeholders.
package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/model"
	"github.com/hacomono-lib/go-i18ngen/internal/parser"
	"github.com/hacomono-lib/go-i18ngen/internal/templatex"
)

func Run(cfg *config.Config) (returnErr error) {
	// Add panic recovery mechanism to prevent unexpected crashes
	defer func() {
		if r := recover(); r != nil {
			returnErr = fmt.Errorf("unexpected panic occurred during generation: %v", r)
		}
	}()

	// Validate input configuration
	if cfg == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate required configuration fields
	if cfg.MessagesGlob == "" {
		return fmt.Errorf("messages glob pattern cannot be empty")
	}
	if cfg.PlaceholdersGlob == "" {
		return fmt.Errorf("placeholders glob pattern cannot be empty")
	}
	if cfg.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}
	if len(cfg.Locales) == 0 {
		return fmt.Errorf("no locales specified in configuration")
	}

	// Check message files exist
	messageFiles, globErr := filepath.Glob(cfg.MessagesGlob)
	if globErr != nil {
		return fmt.Errorf("invalid messages glob pattern %q: %w", cfg.MessagesGlob, globErr)
	}

	if len(messageFiles) == 0 {
		return fmt.Errorf("no message files found matching pattern %q", cfg.MessagesGlob)
	}

	// Parse messages and placeholders with enhanced error context
	messages, err := parser.ParseMessages(cfg.MessagesGlob)
	if err != nil {
		return fmt.Errorf(
			"failed to parse message files from pattern %q:\n  %w\n\nSuggestions:\n"+
				"  - Check that message files exist and have valid YAML syntax\n"+
				"  - Verify glob pattern matches your file structure\n"+
				"  - Ensure templates don't exceed complexity limits",
			cfg.MessagesGlob, err)
	}

	placeholders, err := parser.ParsePlaceholders(cfg.PlaceholdersGlob, cfg.Locales, cfg.Compound)
	if err != nil {
		return fmt.Errorf(
			"failed to parse placeholder files from pattern %q:\n  %w\n\nSuggestions:\n"+
				"  - Check that placeholder files have valid YAML syntax\n"+
				"  - Verify placeholder names are valid Go identifiers\n"+
				"  - Ensure all specified locales (%v) have corresponding values",
			cfg.PlaceholdersGlob, err, cfg.Locales)
	}

	// Validate that we have messages after parsing
	if len(messages) == 0 {
		return fmt.Errorf(
			"no messages found after parsing pattern %q\n\nSuggestions:\n"+
				"  - Check that message files exist in the specified location\n"+
				"  - Verify the glob pattern is correct\n"+
				"  - Ensure message files contain valid message definitions",
			cfg.MessagesGlob)
	}

	defs, err := model.Build(messages, placeholders, cfg.Locales)
	if err != nil {
		return fmt.Errorf(
			"failed to build models from parsed data:\n  %w\n\nSuggestions:\n"+
				"  - Check for placeholder type mismatches\n"+
				"  - Verify all message templates reference valid placeholders\n"+
				"  - Ensure suffix notation is used correctly for multiple instances",
			err)
	}

	if mkdirErr := os.MkdirAll(cfg.OutputDir, 0750); mkdirErr != nil {
		return fmt.Errorf(
			"failed to create output directory %q: %w\n\nSuggestions:\n"+
				"  - Check directory permissions\n"+
				"  - Ensure parent directories exist\n"+
				"  - Verify the path is not read-only",
			cfg.OutputDir, mkdirErr)
	}

	// Determine primary locale (first locale in configuration)
	primaryLocale := "en" // Default fallback
	if len(cfg.Locales) > 0 {
		primaryLocale = cfg.Locales[0]
	}

	// Generate template data with enhanced error context
	messageTemplates, placeholderTemplates, err := model.BuildTemplates(messages, placeholders, cfg.Locales)
	if err != nil {
		return fmt.Errorf(
			"failed to build templates:\n  %w\n\nSuggestions:\n"+
				"  - Check for missing placeholder definitions\n"+
				"  - Verify template syntax is valid\n"+
				"  - Ensure all referenced placeholders exist",
			err)
	}

	// Generate i18n file
	outputFile := filepath.Join(cfg.OutputDir, "i18n.gen.go")
	if err := templatex.Render(
		outputFile,
		cfg.OutputPackage,
		primaryLocale,
		messageTemplates,
		placeholderTemplates,
		defs.Placeholders,
		defs.Messages,
	); err != nil {
		return fmt.Errorf(
			"failed to render generated code to %q:\n  %w\n\nSuggestions:\n"+
				"  - Check output directory permissions\n"+
				"  - Verify package name is valid\n"+
				"  - Ensure templates generate valid Go code\n"+
				"  - Check for disk space availability",
			outputFile, err)
	}

	return nil
}
