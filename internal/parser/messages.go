package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hacomono-lib/go-i18ngen/internal/model"

	"gopkg.in/yaml.v3"
)

// Pre-compiled regular expressions for better performance
var (
	fieldPattern = regexp.MustCompile(`\{\{\s*\.\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)
)

func ParseMessages(pattern string) ([]model.MessageSource, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern for messages %q: %w", pattern, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no message files found matching pattern %q", pattern)
	}

	var results []model.MessageSource
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("failed to open message file %q: %w", file, err)
		}
		defer f.Close()

		ext := filepath.Ext(file)
		data, err := decodeMessageFile(f, ext)
		if err != nil {
			return nil, fmt.Errorf("failed to decode message file %q (ext: %s): %w", file, ext, err)
		}

		for id, localeTemplates := range data {
			// Validate all locales for duplicate placeholders, complexity, and safety
			for locale, template := range localeTemplates {
				if err := validateNoDuplicatePlaceholders(template); err != nil {
					return nil, fmt.Errorf("validation error in message %q (locale: %s) in file %q: %w", id, locale, file, err)
				}
				if err := validateTemplateComplexity(template); err != nil {
					return nil, fmt.Errorf("complexity validation error in message %q (locale: %s) in file %q: %w", id, locale, file, err)
				}
			}

			// Use primary locale (first available) to extract fields
			var primaryTemplate string
			for _, template := range localeTemplates {
				primaryTemplate = template
				break
			}
			fieldInfos := extractFieldInfos(primaryTemplate)

			results = append(results, model.MessageSource{
				ID:         id,
				Templates:  localeTemplates,
				FieldInfos: fieldInfos,
			})
		}
	}
	return results, nil
}

// validateNoDuplicatePlaceholders checks for duplicate placeholders without suffixes
func validateNoDuplicatePlaceholders(template string) error {
	fieldInfos := extractFieldInfos(template)
	fieldCounts := make(map[string]int)

	for _, info := range fieldInfos {
		// Only check fields without suffixes
		if info.Suffix == "" {
			fieldCounts[info.Name]++
		}
	}

	for fieldName, count := range fieldCounts {
		if count > 1 {
			return fmt.Errorf("duplicate placeholder %q found (%d times) - use suffix notation to distinguish multiple instances (e.g., {{.%s:from}} and {{.%s:to}})", fieldName, count, fieldName, fieldName)
		}
	}

	return nil
}

// validateTemplateComplexity checks for overly complex templates
func validateTemplateComplexity(tmpl string) error {
	// Check for excessive nesting depth
	maxDepth := 5
	currentDepth := 0
	maxCurrentDepth := 0

	for _, char := range tmpl {
		switch char {
		case '{':
			currentDepth++
			if currentDepth > maxCurrentDepth {
				maxCurrentDepth = currentDepth
			}
		case '}':
			currentDepth--
		}
	}

	if maxCurrentDepth > maxDepth {
		return fmt.Errorf("template is too complex: nesting depth %d exceeds maximum %d", maxCurrentDepth, maxDepth)
	}

	// Check for excessive template count
	placeholderCount := strings.Count(tmpl, "{{")
	if placeholderCount > 20 {
		return fmt.Errorf("template is too complex: %d placeholders exceed maximum 20", placeholderCount)
	}

	// Check for potential circular references using simple pattern detection
	if err := validateNoCircularPatterns(tmpl); err != nil {
		return err
	}

	return nil
}

// validateNoCircularPatterns checks for potential circular reference patterns
func validateNoCircularPatterns(tmpl string) error {
	// Extract all field references and check for duplicates without suffixes
	matches := fieldPattern.FindAllStringSubmatch(tmpl, -1)

	fieldCounts := make(map[string]int)
	for _, match := range matches {
		if len(match) > 1 {
			fieldName := match[1]
			fieldCounts[fieldName]++
		}
	}

	// Check for potential circular references (same field used multiple times without suffix notation)
	for _, count := range fieldCounts {
		if count > 1 {
			// This might be a circular reference, but it's handled by duplicate validation
			// Just log a warning for now
			continue
		}
	}

	return nil
}

func extractFieldInfos(tmpl string) []model.FieldInfo {
	results := make([]model.FieldInfo, 0)
	remaining := tmpl

	for {
		start := strings.Index(remaining, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(remaining[start:], "}}")
		if end == -1 {
			break
		}

		// Extract the full expression inside {{}}
		expression := strings.TrimSpace(remaining[start+2 : start+end])

		// Check if it starts with . (field reference)
		if strings.HasPrefix(expression, ".") {
			// Remove the leading dot
			fieldExpression := expression[1:]

			// Split by pipe to handle template functions like .entity | title
			parts := strings.Split(fieldExpression, "|")
			fieldPart := strings.TrimSpace(parts[0])

			// Check for suffix notation (field:suffix)
			var fieldName, suffix string
			if colonIndex := strings.Index(fieldPart, ":"); colonIndex != -1 {
				fieldName = strings.TrimSpace(fieldPart[:colonIndex])
				suffix = strings.TrimSpace(fieldPart[colonIndex+1:])
			} else {
				fieldName = fieldPart
			}

			// Only add non-empty fields
			if fieldName != "" {
				results = append(results, model.FieldInfo{
					Name:   fieldName,
					Suffix: suffix,
				})
			}
		}

		remaining = remaining[start+end+2:]
	}

	// Do not sort to preserve field order
	return results
}

func decodeMessageFile(file *os.File, ext string) (map[string]map[string]string, error) {
	// Read file content once
	content, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}

	// First try compound format (map[string]map[string]string)
	var compoundData map[string]map[string]string
	if ext == ".json" {
		if err := json.Unmarshal(content, &compoundData); err == nil {
			// Return compound format as-is
			return compoundData, nil
		}
	} else {
		if err := yaml.Unmarshal(content, &compoundData); err == nil {
			// Return compound format as-is
			return compoundData, nil
		}
	}

	// Fall back to simple format (map[string]string) and convert to compound format
	var data map[string]string
	if ext == ".json" {
		err = json.Unmarshal(content, &data)
	} else {
		err = yaml.Unmarshal(content, &data)
	}
	if err != nil {
		return nil, err
	}

	// Convert simple format to compound format
	result := make(map[string]map[string]string)
	for id, template := range data {
		result[id] = map[string]string{
			"default": template, // Use "default" as locale for simple format
		}
	}
	return result, nil
}
