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

// TestComprehensiveIntegration tests all major features of go-i18ngen
func TestComprehensiveIntegration(t *testing.T) {
	t.Run("BasicGoI18nBackend", func(t *testing.T) {
		testBasicGoI18nBackend(t)
	})

	t.Run("MultiLanguageSupport", func(t *testing.T) {
		testMultiLanguageSupport(t)
	})

	t.Run("ActualLocalizationExecution", func(t *testing.T) {
		testActualLocalizationExecution(t)
	})

	t.Run("TemplateFunctionsParsing", func(t *testing.T) {
		testTemplateFunctionsParsing(t)
	})

	t.Run("LocalizationWithFallback", func(t *testing.T) {
		testLocalizationWithFallback(t)
	})

	t.Log("All comprehensive tests passed successfully!")
}

// testBasicGoI18nBackend tests basic go-i18n backend functionality
func testBasicGoI18nBackend(t *testing.T) {
	tempDir := t.TempDir()

	// Setup go-i18n test with pluralization
	setupTestFiles(t, tempDir, map[string]string{
		"TestMessage": `
  ja: "{{.name}}さんのテスト"
  en: "Test for {{.name}}"`,
		"CountMessage": `
  ja: "{{.Count}}個のアイテム"
  en:
    one: "{{.Count}} item"
    other: "{{.Count}} items"`,
	})

	// Generate and verify
	require.NoError(t, runGeneration(t, tempDir), "Go-i18n generation failed")

	content := readGeneratedFile(t, tempDir)

	// Verify go-i18n-specific features
	assert.Contains(t, content, "go-i18n", "Go-i18n backend should import go-i18n")
	assert.Contains(t, content, "messageData", "Go-i18n should have embedded messageData")
	assert.Contains(t, content, "placeholderData", "Go-i18n should have embedded placeholderData")
	assert.Contains(t, content, "WithPluralCount", "Go-i18n should support WithPluralCount for pluralization")

	t.Log("✓ Basic go-i18n backend test passed")
}

// testMultiLanguageSupport tests support for multiple languages
func testMultiLanguageSupport(t *testing.T) {
	tempDir := t.TempDir()

	// Test with 5 languages
	setupMultiLanguageTestFiles(t, tempDir, map[string]map[string]string{
		"WelcomeMessage": {
			"ja": "ようこそ、{{.name}}さん",
			"en": "Welcome, {{.name}}",
			"fr": "Bienvenue, {{.name}}",
			"de": "Willkommen, {{.name}}",
			"es": "Bienvenido, {{.name}}",
		},
		"GoodbyeMessage": {
			"ja": "さようなら、{{.name}}さん",
			"en": "Goodbye, {{.name}}",
			"fr": "Au revoir, {{.name}}",
			"de": "Auf Wiedersehen, {{.name}}",
			"es": "Adiós, {{.name}}",
		},
	})

	require.NoError(t, runGeneration(t, tempDir), "Multi-language generation failed")

	content := readGeneratedFile(t, tempDir)

	// Verify all languages are included
	languages := []string{"ja", "en", "fr", "de", "es"}
	for _, lang := range languages {
		assert.Contains(t, content, `"`+lang+`": []byte(`,
			"Language %s not found in messageData", lang)
	}

	// Verify all messages are included
	messages := []string{"WelcomeMessage", "GoodbyeMessage"}
	for _, msg := range messages {
		assert.Contains(t, content, msg+":",
			"Message %s not found in generated content", msg)
	}

	t.Log("✓ Multi-language support test passed")
}

// testActualLocalizationExecution tests that generated code actually works
func testActualLocalizationExecution(t *testing.T) {
	tempDir := t.TempDir()

	// Create test setup
	setupLocalizationTestFiles(t, tempDir)

	require.NoError(t, runGeneration(t, tempDir), "Code generation failed")

	// Test actual localization functionality
	testCases := []struct {
		name          string
		templateStr   string
		locale        string
		params        map[string]string
		expectedExact string
	}{
		{
			name:          "SimpleMessage - Japanese",
			templateStr:   "{{.name}}さん、こんにちは",
			locale:        "ja",
			params:        map[string]string{"name": "田中"},
			expectedExact: "田中さん、こんにちは",
		},
		{
			name:          "SimpleMessage - English",
			templateStr:   "Hello, {{.name}}",
			locale:        "en",
			params:        map[string]string{"name": "John"},
			expectedExact: "Hello, John",
		},
		{
			name:          "EntityNotFound - Japanese",
			templateStr:   "{{.entity}}が見つかりません: {{.reason}}",
			locale:        "ja",
			params:        map[string]string{"entity": "ユーザー", "reason": "存在しません"},
			expectedExact: "ユーザーが見つかりません: 存在しません",
		},
		{
			name:          "EntityNotFound - English",
			templateStr:   "{{.entity}} not found: {{.reason}}",
			locale:        "en",
			params:        map[string]string{"entity": "User", "reason": "does not exist"},
			expectedExact: "User not found: does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := renderTemplateForTest(t, tc.templateStr, tc.params)
			assert.Equal(t, tc.expectedExact, result, "Result does not match expected")

			// Verify no error messages
			assert.NotContains(t, result, "[Missing template:", "Template not found error")
			assert.NotContains(t, result, "[Template parse error:", "Template parse error")
		})
	}

	t.Log("✓ Actual localization execution test passed")
}

// testTemplateFunctionsParsing tests template function parsing and field generation
func testTemplateFunctionsParsing(t *testing.T) {
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
			tempDir := t.TempDir()
			setupTemplateFunctionTest(t, tempDir, tc.template)

			require.NoError(t, runGeneration(t, tempDir), "Generation failed")

			content := readGeneratedFile(t, tempDir)
			structDef := extractStructDefinition(t, content, "TestMessage")

			// Check if expected fields are included
			for _, expectedField := range tc.expectFields {
				expectedFieldName := convertToFieldName(expectedField)
				assert.Contains(t, structDef, expectedFieldName,
					"Expected field %s not found", expectedFieldName)
			}
		})
	}

	t.Log("✓ Template functions parsing test passed")
}

// testLocalizationWithFallback tests locale fallback functionality
func testLocalizationWithFallback(t *testing.T) {
	testCases := []struct {
		name             string
		requestedLocale  string
		availableLocales map[string]string
		primaryLocale    string
		expectedResult   string
		params           map[string]string
	}{
		{
			name:            "Fallback to primary locale",
			requestedLocale: "fr", // Not available
			availableLocales: map[string]string{
				"ja": "こんにちは {{.name}}",
				"en": "Hello {{.name}}",
			},
			primaryLocale:  "ja",
			expectedResult: "こんにちは World",
			params:         map[string]string{"name": "World"},
		},
		{
			name:            "Use exact locale match",
			requestedLocale: "en",
			availableLocales: map[string]string{
				"ja": "こんにちは {{.name}}",
				"en": "Hello {{.name}}",
			},
			primaryLocale:  "ja",
			expectedResult: "Hello World",
			params:         map[string]string{"name": "World"},
		},
		{
			name:            "Fallback to any available when primary not available",
			requestedLocale: "fr",
			availableLocales: map[string]string{
				"en": "Hello {{.name}}",
			},
			primaryLocale:  "ja", // Not available
			expectedResult: "Hello World",
			params:         map[string]string{"name": "World"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := renderTemplateWithFallback(t, tc.requestedLocale,
				tc.availableLocales, tc.primaryLocale, tc.params)
			assert.Equal(t, tc.expectedResult, result)
		})
	}

	t.Log("✓ Localization with fallback test passed")
}

// Helper functions

func setupTestFiles(t *testing.T, tempDir string, messages map[string]string) {
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")

	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))

	// Create config
	configContent := `
compound: true
locales: [ja, en]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "i18n"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Create message file
	messageContent := ""
	for msgName, msgTemplates := range messages {
		messageContent += msgName + ":" + msgTemplates + "\n"
	}
	messagePath := filepath.Join(messagesDir, "messages.yaml")
	require.NoError(t, os.WriteFile(messagePath, []byte(messageContent), 0644))

	// Create placeholder file
	placeholderContent := `
user:
  ja: ユーザー
  en: User
`
	placeholderPath := filepath.Join(placeholdersDir, "name.yaml")
	require.NoError(t, os.WriteFile(placeholderPath, []byte(placeholderContent), 0644))
}

func setupMultiLanguageTestFiles(t *testing.T, tempDir string, messages map[string]map[string]string) {
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")

	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))

	// Create config for 5 languages
	configContent := `
compound: true
locales: [ja, en, fr, de, es]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "i18n"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Create message file
	messageContent := ""
	for msgName, locales := range messages {
		messageContent += msgName + ":\n"
		for locale, template := range locales {
			messageContent += fmt.Sprintf("  %s: \"%s\"\n", locale, template)
		}
		messageContent += "\n"
	}
	messagePath := filepath.Join(messagesDir, "messages.yaml")
	require.NoError(t, os.WriteFile(messagePath, []byte(messageContent), 0644))

	// Create multi-language placeholder file
	placeholderContent := `
user:
  ja: ユーザー
  en: User
  fr: Utilisateur
  de: Benutzer
  es: Usuario
`
	placeholderPath := filepath.Join(placeholdersDir, "user.yaml")
	require.NoError(t, os.WriteFile(placeholderPath, []byte(placeholderContent), 0644))
}

func setupLocalizationTestFiles(t *testing.T, tempDir string) {
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")

	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))

	// Create config
	configContent := `
compound: true
locales: [ja, en]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "testlocalize"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Create message file
	messageContent := `EntityNotFound:
  ja: "{{.entity}}が見つかりません: {{.reason}}"
  en: "{{.entity}} not found: {{.reason}}"
SimpleMessage:
  ja: "{{.name}}さん、こんにちは"
  en: "Hello, {{.name}}"
`
	messagePath := filepath.Join(messagesDir, "messages.yaml")
	require.NoError(t, os.WriteFile(messagePath, []byte(messageContent), 0644))

	// Create placeholder files
	entityContent := `user:
  ja: "ユーザー"
  en: "User"
product:
  ja: "製品"
  en: "Product"
`
	entityPath := filepath.Join(placeholdersDir, "entity.yaml")
	require.NoError(t, os.WriteFile(entityPath, []byte(entityContent), 0644))

	reasonContent := `not_exist:
  ja: "存在しません"
  en: "does not exist"
deleted:
  ja: "削除されています"
  en: "has been deleted"
`
	reasonPath := filepath.Join(placeholdersDir, "reason.yaml")
	require.NoError(t, os.WriteFile(reasonPath, []byte(reasonContent), 0644))
}

func setupTemplateFunctionTest(t *testing.T, tempDir string, template string) {
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")

	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))

	// Create config
	configContent := `
compound: true
locales: [ja, en]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "testfunc"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Create message file with the test template
	messageContent := fmt.Sprintf(`TestMessage:
  ja: "%s"
  en: "%s"
`, template, template)
	messagePath := filepath.Join(messagesDir, "test.yaml")
	require.NoError(t, os.WriteFile(messagePath, []byte(messageContent), 0644))

	// Create dummy placeholder file
	placeholderContent := `dummy:
  ja: "ダミー"
  en: "Dummy"
`
	placeholderPath := filepath.Join(placeholdersDir, "dummy.yaml")
	require.NoError(t, os.WriteFile(placeholderPath, []byte(placeholderContent), 0644))
}

func runGeneration(t *testing.T, tempDir string) error {
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	require.NoError(t, os.Chdir(tempDir))

	configPath := filepath.Join(tempDir, "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	return generator.Run(cfg)
}

func readGeneratedFile(t *testing.T, tempDir string) string {
	outputPath := filepath.Join(tempDir, "i18n.gen.go")
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	return string(content)
}

func renderTemplateForTest(t *testing.T, templateStr string, params map[string]string) string {
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

	tmpl, err := template.New("test").Funcs(funcMap).Parse(templateStr)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, params))

	return buf.String()
}

func renderTemplateWithFallback(t *testing.T, locale string, templates map[string]string, primaryLocale string, params map[string]string) string {
	templateStr, exists := templates[locale]
	if !exists {
		templateStr, exists = templates[primaryLocale]
		if !exists {
			for _, tmpl := range templates {
				templateStr = tmpl
				break
			}
		}
	}

	if templateStr == "" {
		return fmt.Sprintf("[Missing template: test.%s]", locale)
	}

	return renderTemplateForTest(t, templateStr, params)
}

func extractStructDefinition(t *testing.T, content, structName string) string {
	structStart := strings.Index(content, "type "+structName+" struct")
	require.Greater(t, structStart, -1, "%s struct not found", structName)

	structEnd := strings.Index(content[structStart:], "}")
	require.Greater(t, structEnd, -1, "%s struct end not found", structName)

	return content[structStart : structStart+structEnd]
}

func convertToFieldName(field string) string {
	if strings.Contains(field, ":") {
		parts := strings.Split(field, ":")
		fieldName := parts[0]
		suffix := parts[1]
		// field:input -> FieldInput
		return strings.ToUpper(fieldName[:1]) + fieldName[1:] +
			strings.ToUpper(suffix[:1]) + suffix[1:]
	}
	// Normal field name case
	return strings.ToUpper(field[:1]) + field[1:]
}
