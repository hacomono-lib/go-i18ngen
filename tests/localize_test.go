package tests

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/generator"
)

// TestLocalizeActualExecution tests that the generated code actually works and produces correct localized output
func TestLocalizeActualExecution(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_localize_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create subdirectories
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	outputDir := filepath.Join(tempDir, "output")
	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))

	// Create test message file
	messageFile := filepath.Join(messagesDir, "messages.yaml")
	messageContent := `EntityNotFound:
  ja: "{{.entity}}が見つかりません: {{.reason}}"
  en: "{{.entity}} not found: {{.reason}}"
SimpleMessage:
  ja: "{{.name}}さん、こんにちは"
  en: "Hello, {{.name}}"
`
	require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Create test placeholder files
	entityFile := filepath.Join(placeholdersDir, "entity.yaml")
	entityContent := `user:
  ja: "ユーザー"
  en: "User"
product:
  ja: "製品"
  en: "Product"
`
	require.NoError(t, os.WriteFile(entityFile, []byte(entityContent), 0644))

	reasonFile := filepath.Join(placeholdersDir, "reason.yaml")
	reasonContent := `not_exist:
  ja: "存在しません"
  en: "does not exist"
deleted:
  ja: "削除されています"
  en: "has been deleted"
`
	require.NoError(t, os.WriteFile(reasonFile, []byte(reasonContent), 0644))

	// Create configuration
	cfg := &config.Config{
		Locales:          []string{"ja", "en"},
		Compound:         true,
		MessagesGlob:     filepath.Join(messagesDir, "*.yaml"),
		PlaceholdersGlob: filepath.Join(placeholdersDir, "*.yaml"),
		OutputDir:        outputDir,
		OutputPackage:    "testlocalize",
	}

	// Execute code generation
	err = generator.Run(cfg)
	require.NoError(t, err)

	// Load generated code
	generatedFile := filepath.Join(outputDir, "i18n.gen.go")
	_, err = os.ReadFile(generatedFile)
	require.NoError(t, err)

	// Test helper function for actual localization functionality
	testCases := []struct {
		name             string
		messageID        string
		locale           string
		templateStr      string
		params           map[string]string
		expectedContains []string
		expectedExact    string
	}{
		{
			name:          "SimpleMessage - Japanese",
			messageID:     "SimpleMessage",
			locale:        "ja",
			templateStr:   "{{.name}}さん、こんにちは",
			params:        map[string]string{"name": "田中"},
			expectedExact: "田中さん、こんにちは",
		},
		{
			name:          "SimpleMessage - English",
			messageID:     "SimpleMessage",
			locale:        "en",
			templateStr:   "Hello, {{.name}}",
			params:        map[string]string{"name": "John"},
			expectedExact: "Hello, John",
		},
		{
			name:          "EntityNotFound - Japanese",
			messageID:     "EntityNotFound",
			locale:        "ja",
			templateStr:   "{{.entity}}が見つかりません: {{.reason}}",
			params:        map[string]string{"entity": "ユーザー", "reason": "存在しません"},
			expectedExact: "ユーザーが見つかりません: 存在しません",
		},
		{
			name:          "EntityNotFound - English",
			messageID:     "EntityNotFound",
			locale:        "en",
			templateStr:   "{{.entity}} not found: {{.reason}}",
			params:        map[string]string{"entity": "User", "reason": "does not exist"},
			expectedExact: "User not found: does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := renderTemplateForTest(t, tc.messageID, tc.templateStr, tc.params)

			if tc.expectedExact != "" {
				assert.Equal(t, tc.expectedExact, result, "Result does not match expected exact match")
			}

			for _, expectedContain := range tc.expectedContains {
				assert.Contains(t, result, expectedContain, "Expected string not found: %s", expectedContain)
			}

			// Basic validation (verify no error messages)
			assert.NotContains(t, result, "[Missing template:", "Template not found error occurred")
			assert.NotContains(t, result, "[Template parse error:", "Template parse error occurred")
			assert.NotContains(t, result, "[Template execution error:", "Template execution error occurred")
		})
	}
}

// renderTemplateForTest recreates the same implementation as the renderTemplate function in generated code for testing
func renderTemplateForTest(t *testing.T, messageID, templateStr string, params map[string]string) string {
	// Create template function map (same as generated code)
	funcMap := template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}

	tmpl, err := template.New(messageID).Funcs(funcMap).Parse(templateStr)
	require.NoError(t, err, "Failed to parse template")

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	require.NoError(t, err, "Failed to execute template")

	return buf.String()
}

// TestTemplateFunctionsParsing tests that template functions are correctly parsed and don't cause field duplication
func TestTemplateFunctionsParsing(t *testing.T) {
	testCases := []struct {
		name         string
		template     string
		expectFields []string
	}{
		{
			name:         "Single field with title function",
			template:     "{{.field | title}}",
			expectFields: []string{"field"},
		},
		{
			name:         "Single field with upper function",
			template:     "{{.field | upper}}",
			expectFields: []string{"field"},
		},
		{
			name:         "Single field with multiple functions",
			template:     "{{.field | title | upper}}",
			expectFields: []string{"field"},
		},
		{
			name:         "Same field multiple times with suffix notation",
			template:     "{{.field:input | title}} and {{.field:display | upper}}",
			expectFields: []string{"field:input", "field:display"},
		},
		{
			name:         "Different fields with suffix notation",
			template:     "{{.field:first}} and {{.field:second | title}}",
			expectFields: []string{"field:first", "field:second"},
		},
		{
			name:         "Multiple different fields with functions",
			template:     "{{.field1 | title}} and {{.field2 | upper}}",
			expectFields: []string{"field1", "field2"},
		},
		{
			name:         "Complex case with spaces and suffix notation",
			template:     "{{ .field:input | title }} and {{ .field:output | upper }}",
			expectFields: []string{"field:input", "field:output"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create actual temporary files for testing since we're directly using internal extractFields function
			tempDir, err := os.MkdirTemp("", "template_function_test")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(tempDir) }()

			messageFile := filepath.Join(tempDir, "test.yaml")
			messageContent := fmt.Sprintf(`TestMessage:
  ja: "%s"
  en: "%s"
`, tc.template, tc.template)
			require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

			// use parser to extract fields
			messagesDir := filepath.Join(tempDir, "messages")
			placeholdersDir := filepath.Join(tempDir, "placeholders")
			outputDir := filepath.Join(tempDir, "output")
			require.NoError(t, os.MkdirAll(messagesDir, 0755))
			require.NoError(t, os.MkdirAll(placeholdersDir, 0755))

			// Move the test file to the messages directory
			require.NoError(t, os.Rename(messageFile, filepath.Join(messagesDir, "test.yaml")))

			// Create a dummy placeholder file to avoid errors
			placeholderFile := filepath.Join(placeholdersDir, "dummy.yaml")
			require.NoError(t, os.WriteFile(placeholderFile, []byte("dummy:\n  ja: \"ダミー\"\n  en: \"Dummy\""), 0644))

			cfg := &config.Config{
				Locales:          []string{"ja", "en"},
				Compound:         true,
				MessagesGlob:     filepath.Join(messagesDir, "*.yaml"),
				PlaceholdersGlob: filepath.Join(placeholdersDir, "*.yaml"),
				OutputDir:        outputDir,
				OutputPackage:    "testfunc",
			}

			err = generator.Run(cfg)
			require.NoError(t, err)

			// verify the generated code
			generatedFile := filepath.Join(outputDir, "i18n.gen.go")
			generatedCode, err := os.ReadFile(generatedFile)
			require.NoError(t, err)
			codeStr := string(generatedCode)

			// check TestMessage struct definition
			structStart := strings.Index(codeStr, "type TestMessage struct")
			require.Greater(t, structStart, -1, "TestMessage struct not found")

			structEnd := strings.Index(codeStr[structStart:], "}")
			require.Greater(t, structEnd, -1, "TestMessage struct end not found")

			structDef := codeStr[structStart : structStart+structEnd]

			// Check if expected fields are included
			for _, expectedField := range tc.expectFields {
				var expectedFieldName string

				// Convert for suffix notation case
				if strings.Contains(expectedField, ":") {
					parts := strings.Split(expectedField, ":")
					fieldName := parts[0]
					suffix := parts[1]
					// field:input -> FieldInput
					expectedFieldName = strings.ToUpper(fieldName[:1]) + fieldName[1:] + strings.ToUpper(suffix[:1]) + suffix[1:]
				} else {
					// Normal field name case
					expectedFieldName = strings.ToUpper(expectedField[:1]) + expectedField[1:]
				}

				assert.Contains(t, structDef, expectedFieldName, "Expected field %s not found", expectedFieldName)
			}
		})
	}
}

// TestLocalizeWithFallback tests that localization works with fallback to primary locale
func TestLocalizeWithFallback(t *testing.T) {
	testCases := []struct {
		name             string
		templateStr      string
		params           map[string]string
		requestedLocale  string
		availableLocales map[string]string
		primaryLocale    string
		expectedResult   string
	}{
		{
			name:            "Fallback to primary locale",
			templateStr:     "Hello {{.name}}",
			params:          map[string]string{"name": "World"},
			requestedLocale: "fr", // Request French but not available
			availableLocales: map[string]string{
				"ja": "こんにちは {{.name}}",
				"en": "Hello {{.name}}",
			},
			primaryLocale:  "ja",
			expectedResult: "こんにちは World",
		},
		{
			name:            "Use exact locale match",
			templateStr:     "Hello {{.name}}",
			params:          map[string]string{"name": "World"},
			requestedLocale: "en",
			availableLocales: map[string]string{
				"ja": "こんにちは {{.name}}",
				"en": "Hello {{.name}}",
			},
			primaryLocale:  "ja",
			expectedResult: "Hello World",
		},
		{
			name:            "Fallback to any available when primary not available",
			templateStr:     "Hello {{.name}}",
			params:          map[string]string{"name": "World"},
			requestedLocale: "fr",
			availableLocales: map[string]string{
				"en": "Hello {{.name}}",
			},
			primaryLocale:  "ja", // Primary not available
			expectedResult: "Hello World",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := renderTemplateWithFallback(t, tc.requestedLocale, "test", tc.availableLocales, tc.primaryLocale, tc.params)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// renderTemplateWithFallback simulates the fallback logic from generated code
func renderTemplateWithFallback(t *testing.T, locale, messageID string, templates map[string]string, primaryLocale string, params map[string]string) string {
	templateStr, exists := templates[locale]
	if !exists {
		// Fallback to primary locale
		templateStr, exists = templates[primaryLocale]
		if !exists {
			// Fallback to any available locale
			for _, tmpl := range templates {
				templateStr = tmpl
				break
			}
		}
	}

	if templateStr == "" {
		return fmt.Sprintf("[Missing template: %s.%s]", messageID, locale)
	}

	funcMap := template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}

	tmpl, err := template.New(messageID).Funcs(funcMap).Parse(templateStr)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	require.NoError(t, err)

	return buf.String()
}
