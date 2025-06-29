// Package model defines the data structures and logic for building message and placeholder definitions.
package model

import (
	"fmt"
	"regexp"
	"strings"
)

// Pre-compiled regular expressions for better performance
var (
	templateFieldPattern       = regexp.MustCompile(`\{\{\s*\.\s*([a-zA-Z_][a-zA-Z0-9_]*)(\s*\|[^}]*)?\s*\}\}`)
	templateFieldSuffixPattern = regexp.MustCompile(`\{\{\s*\.\s*([a-zA-Z_][a-zA-Z0-9_]*(?::[a-zA-Z0-9_]+)?)(\s*\|[^}]*)?\s*\}\}`)
)

// processTemplateForDuplicates converts template strings to use numbered placeholders for duplicates
// Example: "{{.name}} hello, {{.name}} world" -> "{{.name1}} hello, {{.name2}} world"
func processTemplateForDuplicates(template string, fields []string) string {
	// Count occurrences of each field in the original fields list
	fieldCounts := make(map[string]int)
	for _, f := range fields {
		fieldCounts[f]++
	}

	// Track current index for each field as we process the template
	fieldIndices := make(map[string]int)

	// Process the template string
	result := template

	// Find all {{.field}} patterns
	// Replace placeholders with numbered versions for duplicates
	result = templateFieldPattern.ReplaceAllStringFunc(result, func(match string) string {
		// Extract the field name and any template functions
		submatches := templateFieldPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		fieldName := submatches[1]
		templateFunctions := ""
		if len(submatches) > 2 {
			templateFunctions = submatches[2]
		}

		// Check if this field has duplicates
		if fieldCounts[fieldName] > 1 {
			fieldIndices[fieldName]++
			return fmt.Sprintf("{{.%s%d%s}}", fieldName, fieldIndices[fieldName], templateFunctions)
		}

		return match
	})

	return result
}

// ProcessMessageTemplates processes all templates in a message to handle duplicate placeholders
func ProcessMessageTemplates(templates map[string]string, fields []string) map[string]string {
	result := make(map[string]string)
	for locale, template := range templates {
		result[locale] = processTemplateForDuplicates(template, fields)
	}
	return result
}

// ProcessMessageTemplatesWithFieldInfos processes templates using FieldInfo for suffix-based placeholders
func ProcessMessageTemplatesWithFieldInfos(templates map[string]string, fieldInfos []FieldInfo) map[string]string {
	result := make(map[string]string)
	for locale, template := range templates {
		result[locale] = processTemplateWithFieldInfos(template, fieldInfos)
	}
	return result
}

// processTemplateWithFieldInfos converts template strings to use suffix-based placeholders
// Example: "{{.entity:from}} to {{.entity:to}}" -> "{{.entityFrom}} to {{.entityTo}}"
func processTemplateWithFieldInfos(template string, fieldInfos []FieldInfo) string {
	result := template

	// Find all {{.field}} patterns and replace with appropriate keys
	// Replace placeholders with template keys
	result = templateFieldSuffixPattern.ReplaceAllStringFunc(result, func(match string) string {
		// Extract the field name and any template functions
		submatches := templateFieldSuffixPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		fieldExpression := submatches[1]
		templateFunctions := ""
		if len(submatches) > 2 {
			templateFunctions = submatches[2]
		}

		// Find matching FieldInfo for this expression
		for _, fieldInfo := range fieldInfos {
			if fieldInfo.String() == fieldExpression {
				templateKey := fieldInfo.GenerateTemplateKey()
				return fmt.Sprintf("{{.%s%s}}", templateKey, templateFunctions)
			}
		}

		// If no match found, return original (should not happen in normal cases)
		return match
	})

	return result
}

// extractTemplateFunctions extracts template functions from a template field expression
// Example: "{{.field:input | title | upper}}" -> []string{"title", "upper"}
func extractTemplateFunctions(templateFunctions string) []string {
	if templateFunctions == "" {
		return nil
	}
	
	// Remove leading pipe and whitespace
	functions := strings.TrimSpace(strings.TrimPrefix(templateFunctions, "|"))
	
	// Split by pipe and clean up
	parts := strings.Split(functions, "|")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}

// BuildTemplateFunctionsMetadata builds template function metadata for go-i18n backend
func BuildTemplateFunctionsMetadata(messages []MessageSource, locales []string) map[string]map[string]map[string][]string {
	metadata := make(map[string]map[string]map[string][]string)
	
	for _, msg := range messages {
		metadata[msg.ID] = make(map[string]map[string][]string)
		
		for _, locale := range locales {
			metadata[msg.ID][locale] = make(map[string][]string)
			
			template, exists := msg.Templates[locale]
			if !exists {
				continue
			}
			
			// Extract template functions for each field
			for _, fieldInfo := range msg.FieldInfos {
				functions := extractTemplateFunctionsFromTemplate(template, fieldInfo)
				if len(functions) > 0 {
					metadata[msg.ID][locale][fieldInfo.GenerateTemplateKey()] = functions
				}
			}
		}
	}
	
	return metadata
}

// extractTemplateFunctionsFromTemplate extracts template functions for a specific field from a template
func extractTemplateFunctionsFromTemplate(template string, fieldInfo FieldInfo) []string {
	// Create pattern to match the specific field with functions
	pattern := fmt.Sprintf(`\{\{\s*\.%s(\s*\|[^}]*)?\s*\}\}`, regexp.QuoteMeta(fieldInfo.String()))
	re := regexp.MustCompile(pattern)
	
	matches := re.FindStringSubmatch(template)
	if len(matches) < 2 {
		return nil
	}
	
	return extractTemplateFunctions(matches[1])
}
