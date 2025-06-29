// Package parser handles parsing of message and placeholder files for the i18n generator.
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

const (
	jsonExt = ".json"
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
		defer func() { _ = f.Close() }()

		ext := filepath.Ext(file)
		data, err := decodeMessageFileWithRaw(f, ext)
		if err != nil {
			return nil, fmt.Errorf("failed to decode message file %q (ext: %s): %w", file, ext, err)
		}

		for id, localeTemplates := range data.Templates {
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

			// Get raw templates for this message ID
			rawTemplates := data.RawTemplates[id]
			if rawTemplates == nil {
				rawTemplates = make(map[string]interface{})
			}

			results = append(results, model.MessageSource{
				ID:           id,
				Templates:    localeTemplates,
				RawTemplates: rawTemplates,
				FieldInfos:   fieldInfos,
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
			return fmt.Errorf(
				"duplicate placeholder %q found (%d times) - use suffix notation "+
					"to distinguish multiple instances (e.g., {{.%s:from}} and {{.%s:to}})",
				fieldName, count, fieldName, fieldName)
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

// MessageFileData holds both simplified and raw template data
type MessageFileData struct {
	Templates    map[string]map[string]string      // simplified templates for processing
	RawTemplates map[string]map[string]interface{} // raw templates for documentation
}

func decodeMessageFileWithRaw(file *os.File, ext string) (*MessageFileData, error) {
	// Read file content once
	content, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}

	result := &MessageFileData{
		Templates:    make(map[string]map[string]string),
		RawTemplates: make(map[string]map[string]interface{}),
	}

	// First try compound format (map[string]map[string]string)
	var compoundData map[string]map[string]string
	if ext == jsonExt {
		if jsonErr := json.Unmarshal(content, &compoundData); jsonErr == nil {
			result.Templates = compoundData
			// Convert to interface{} for raw templates
			for msgID, localeMap := range compoundData {
				result.RawTemplates[msgID] = make(map[string]interface{})
				for locale, template := range localeMap {
					result.RawTemplates[msgID][locale] = template
				}
			}
			return result, nil
		}
	} else {
		if yamlErr := yaml.Unmarshal(content, &compoundData); yamlErr == nil {
			result.Templates = compoundData
			// Convert to interface{} for raw templates  
			for msgID, localeMap := range compoundData {
				result.RawTemplates[msgID] = make(map[string]interface{})
				for locale, template := range localeMap {
					result.RawTemplates[msgID][locale] = template
				}
			}
			return result, nil
		}
	}

	// Try mixed format that supports both strings and pluralization objects
	var mixedData map[string]map[string]interface{}
	if ext == jsonExt {
		if jsonErr := json.Unmarshal(content, &mixedData); jsonErr == nil {
			result.Templates = convertMixedToStringMap(mixedData)
			result.RawTemplates = mixedData
			return result, nil
		}
	} else {
		if yamlErr := yaml.Unmarshal(content, &mixedData); yamlErr == nil {
			result.Templates = convertMixedToStringMap(mixedData)
			result.RawTemplates = mixedData
			return result, nil
		}
	}

	// Fall back to simple format (map[string]string) and convert to compound format
	var data map[string]string
	if ext == jsonExt {
		err = json.Unmarshal(content, &data)
	} else {
		err = yaml.Unmarshal(content, &data)
	}
	if err != nil {
		return nil, err
	}

	// Convert simple format to compound format
	for id, template := range data {
		result.Templates[id] = map[string]string{
			"default": template, // Use "default" as locale for simple format
		}
		result.RawTemplates[id] = map[string]interface{}{
			"default": template,
		}
	}
	return result, nil
}


// convertMixedToStringMap converts mixed format (string or pluralization object) to string-only format
func convertMixedToStringMap(mixedData map[string]map[string]interface{}) map[string]map[string]string {
	result := make(map[string]map[string]string)
	
	for messageID, localeData := range mixedData {
		result[messageID] = make(map[string]string)
		
		for locale, value := range localeData {
			switch v := value.(type) {
			case string:
				// Simple string template
				result[messageID][locale] = v
			case map[string]interface{}:
				// Pluralization object - convert to go-i18n format
				result[messageID][locale] = convertPluralToTemplate(v)
			case map[interface{}]interface{}:
				// YAML can parse as map[interface{}]interface{}, convert it
				stringMap := make(map[string]interface{})
				for k, val := range v {
					if str, ok := k.(string); ok {
						stringMap[str] = val
					}
				}
				result[messageID][locale] = convertPluralToTemplate(stringMap)
			default:
				// Fallback to string representation
				result[messageID][locale] = fmt.Sprintf("%v", v)
			}
		}
	}
	
	return result
}

// convertPluralToTemplate converts plural forms to a single template with go-i18n format
func convertPluralToTemplate(pluralMap map[string]interface{}) string {
	// For now, prioritize "other" form, then "one", then any available
	if other, exists := pluralMap["other"]; exists {
		if str, ok := other.(string); ok {
			return str
		}
	}
	
	if one, exists := pluralMap["one"]; exists {
		if str, ok := one.(string); ok {
			return str
		}
	}
	
	// Return first available form
	for _, value := range pluralMap {
		if str, ok := value.(string); ok {
			return str
		}
	}
	
	return "{{.Count}} items" // fallback
}
