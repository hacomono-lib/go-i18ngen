package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFieldInfos(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected []model.FieldInfo
	}{
		{
			name:     "simple placeholder",
			template: "{{.entity}}",
			expected: []model.FieldInfo{{Name: "entity", Suffix: ""}},
		},
		{
			name:     "multiple placeholders",
			template: "{{.entity}} and {{.reason}}",
			expected: []model.FieldInfo{{Name: "entity", Suffix: ""}, {Name: "reason", Suffix: ""}},
		},
		{
			name:     "placeholder with template function",
			template: "{{.entity | title}}",
			expected: []model.FieldInfo{{Name: "entity", Suffix: ""}},
		},
		{
			name:     "placeholder with suffix notation",
			template: "{{.entity:from}}",
			expected: []model.FieldInfo{{Name: "entity", Suffix: "from"}},
		},
		{
			name:     "multiple suffix placeholders",
			template: "{{.entity:from}} to {{.entity:to}}",
			expected: []model.FieldInfo{{Name: "entity", Suffix: "from"}, {Name: "entity", Suffix: "to"}},
		},
		{
			name:     "suffix with template function",
			template: "{{.entity:from | title}}",
			expected: []model.FieldInfo{{Name: "entity", Suffix: "from"}},
		},
		{
			name:     "mixed case with suffix notation",
			template: "Please {{.action}} the {{.entity:from}} to {{.entity:to | title}}",
			expected: []model.FieldInfo{{Name: "action", Suffix: ""}, {Name: "entity", Suffix: "from"}, {Name: "entity", Suffix: "to"}},
		},
		{
			name:     "complex template functions with suffix notation",
			template: "Hello {{.name:user | title | upper}}, welcome to {{.name:owner}}'s account",
			expected: []model.FieldInfo{{Name: "name", Suffix: "user"}, {Name: "name", Suffix: "owner"}},
		},
		{
			name:     "empty template",
			template: "",
			expected: []model.FieldInfo{},
		},
		{
			name:     "no placeholders",
			template: "Simple message",
			expected: []model.FieldInfo{},
		},
		{
			name:     "incomplete placeholder",
			template: "{{.entity not found",
			expected: []model.FieldInfo{},
		},
		{
			name:     "suffix notation single",
			template: "{{.entity:from}}",
			expected: []model.FieldInfo{{Name: "entity", Suffix: "from"}},
		},
		{
			name:     "suffix notation multiple",
			template: "{{.entity:from}} to {{.entity:to}}",
			expected: []model.FieldInfo{{Name: "entity", Suffix: "from"}, {Name: "entity", Suffix: "to"}},
		},
		{
			name:     "suffix notation with template functions",
			template: "{{.entity:from | title}} to {{.entity:to | upper}}",
			expected: []model.FieldInfo{{Name: "entity", Suffix: "from"}, {Name: "entity", Suffix: "to"}},
		},
		{
			name:     "mixed suffix and regular placeholders",
			template: "{{.name}} moved {{.entity:from}} to {{.entity:to}}",
			expected: []model.FieldInfo{{Name: "name", Suffix: ""}, {Name: "entity", Suffix: "from"}, {Name: "entity", Suffix: "to"}},
		},
		{
			name:     "numeric suffix",
			template: "{{.item:1}} and {{.item:2}}",
			expected: []model.FieldInfo{{Name: "item", Suffix: "1"}, {Name: "item", Suffix: "2"}},
		},
		{
			name:     "complex suffix names",
			template: "{{.field:input_value}} to {{.field:display_name}}",
			expected: []model.FieldInfo{{Name: "field", Suffix: "input_value"}, {Name: "field", Suffix: "display_name"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFieldInfos(tt.template)
			assert.Equal(t, tt.expected, result, "Field extraction does not match expected values")
		})
	}
}

func TestParseMessages(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test message file with only valid syntax (no duplicate placeholders)
	messageFile := filepath.Join(tempDir, "messages.yaml")
	messageContent := `EntityNotFound:
  ja: "{{.entity}}が見つかりません: {{.reason}}"
  en: "{{.entity}} not found: {{.reason}}"
SuffixExample:
  ja: "{{.name:user}}さん、{{.name:owner}}さんのアカウントへようこそ"
  en: "Welcome {{.name:user}}, to {{.name:owner}}'s account"
TemplateFunctionExample:
  ja: "{{.field:input}}に{{.field:display | upper}}エラー"
  en: "{{.field:input | title}} error in {{.field:display}}"
`
	require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Execute ParseMessages
	pattern := filepath.Join(tempDir, "*.yaml")
	results, err := ParseMessages(pattern)
	require.NoError(t, err)

	// Verify results
	assert.Len(t, results, 3, "Number of messages does not match expected")

	// Verify EntityNotFound
	entityNotFound := findMessageByID(results, "EntityNotFound")
	require.NotNil(t, entityNotFound, "EntityNotFound message not found")

	expectedEntityFields := []model.FieldInfo{{Name: "entity", Suffix: ""}, {Name: "reason", Suffix: ""}}
	assert.Equal(t, expectedEntityFields, entityNotFound.FieldInfos)
	assert.Equal(t, "{{.entity}}が見つかりません: {{.reason}}", entityNotFound.Templates["ja"])
	assert.Equal(t, "{{.entity}} not found: {{.reason}}", entityNotFound.Templates["en"])

	// Verify SuffixExample (suffix notation)
	suffixExample := findMessageByID(results, "SuffixExample")
	require.NotNil(t, suffixExample, "SuffixExample message not found")

	expectedSuffixFields := []model.FieldInfo{{Name: "name", Suffix: "user"}, {Name: "name", Suffix: "owner"}}
	assert.Equal(t, expectedSuffixFields, suffixExample.FieldInfos, "Suffix notation placeholders are not properly processed")

	// Verify TemplateFunctionExample
	templateFunctionExample := findMessageByID(results, "TemplateFunctionExample")
	require.NotNil(t, templateFunctionExample, "TemplateFunctionExample message not found")

	expectedTemplateFields := []model.FieldInfo{{Name: "field", Suffix: "input"}, {Name: "field", Suffix: "display"}}
	assert.Equal(t, expectedTemplateFields, templateFunctionExample.FieldInfos, "Placeholders with template functions are not properly processed")
}

func TestParseMessagesWithJSON(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_json_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create JSON format test message file with suffix notation
	messageFile := filepath.Join(tempDir, "messages.json")
	messageContent := `{
  "ValidationError": {
    "ja": "{{.field:input}}の{{.field:display | upper}}検証エラー",
    "en": "{{.field:input | title}} validation error for {{.field:display}}"
  }
}`
	require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Execute ParseMessages
	pattern := filepath.Join(tempDir, "*.json")
	results, err := ParseMessages(pattern)
	require.NoError(t, err)

	// Verify results
	assert.Len(t, results, 1)
	validationError := results[0]
	assert.Equal(t, "ValidationError", validationError.ID)

	expectedFields := []model.FieldInfo{{Name: "field", Suffix: "input"}, {Name: "field", Suffix: "display"}}
	assert.Equal(t, expectedFields, validationError.FieldInfos, "Verify that suffix notation and template function processing work with JSON format")
}

func TestParseMessagesDuplicatePlaceholderValidation(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_validation_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test message file with duplicate placeholders (should fail)
	messageFile := filepath.Join(tempDir, "messages.yaml")
	messageContent := `InvalidMessage:
  ja: "{{.name}}さん、{{.name}}さんのアカウントへようこそ"
  en: "Welcome {{.name}}, to {{.name}}'s account"
`
	require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Execute ParseMessages - should return error
	pattern := filepath.Join(tempDir, "*.yaml")
	results, err := ParseMessages(pattern)
	assert.Error(t, err, "Should return error for duplicate placeholders")
	assert.Contains(t, err.Error(), "duplicate placeholder", "Error message should mention duplicate placeholder")
	assert.Contains(t, err.Error(), "suffix notation", "Error message should suggest suffix notation")
	assert.Nil(t, results)
}

func TestParseMessagesEmptyPattern(t *testing.T) {
	// Test with non-existent pattern
	results, err := ParseMessages("/nonexistent/*.yaml")
	assert.Error(t, err, "Should return error for non-existent patterns")
	assert.Contains(t, err.Error(), "no message files found", "Error should indicate no files found")
	assert.Nil(t, results)
}

func TestDecodeMessageFileErrors(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_error_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create invalid YAML file
	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	invalidContent := `invalid: yaml: content:
  - unclosed
    brackets: [`
	require.NoError(t, os.WriteFile(invalidFile, []byte(invalidContent), 0644))

	// Verify that error is returned
	pattern := filepath.Join(tempDir, "*.yaml")
	results, err := ParseMessages(pattern)
	assert.Error(t, err, "Verify that error is returned for invalid YAML files")
	assert.Nil(t, results)
}

// Helper function
func findMessageByID(messages []model.MessageSource, id string) *model.MessageSource {
	for i := range messages {
		if messages[i].ID == id {
			return &messages[i]
		}
	}
	return nil
}

// ParsePlaceholders tests
func TestParsePlaceholders(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_placeholder_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test placeholder files
	// Simple format files
	fieldFile := filepath.Join(tempDir, "field.ja.yaml")
	fieldContent := `EmailAddress: "メールアドレス"
FirstName: "名前"
LastName: "苗字"`
	require.NoError(t, os.WriteFile(fieldFile, []byte(fieldContent), 0644))

	fieldEnFile := filepath.Join(tempDir, "field.en.yaml")
	fieldEnContent := `EmailAddress: "Email Address"
FirstName: "First Name"
LastName: "Last Name"`
	require.NoError(t, os.WriteFile(fieldEnFile, []byte(fieldEnContent), 0644))

	// Execute ParsePlaceholders
	pattern := filepath.Join(tempDir, "*.yaml")
	locales := []string{"ja", "en"}
	results, err := ParsePlaceholders(pattern, locales, false)
	require.NoError(t, err)

	// Verify results
	assert.Len(t, results, 1, "Should have one placeholder source")
	assert.Equal(t, "field", results[0].Kind)
	assert.Len(t, results[0].Items, 3, "Should have three items")

	// Verify specific items
	assert.Contains(t, results[0].Items, "EmailAddress")
	assert.Contains(t, results[0].Items, "FirstName")
	assert.Contains(t, results[0].Items, "LastName")

	// Verify locales
	emailItem := results[0].Items["EmailAddress"]
	assert.Equal(t, "メールアドレス", emailItem["ja"])
	assert.Equal(t, "Email Address", emailItem["en"])
}

func TestParsePlaceholdersCompoundFormat(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_compound_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create compound format file
	compoundFile := filepath.Join(tempDir, "validation.yaml")
	compoundContent := `EmailAddress:
  ja: "メールアドレス"
  en: "Email Address"
Required:
  ja: "必須"
  en: "Required"`
	require.NoError(t, os.WriteFile(compoundFile, []byte(compoundContent), 0644))

	// Execute ParsePlaceholders with compound format
	pattern := filepath.Join(tempDir, "*.yaml")
	locales := []string{"ja", "en"}
	results, err := ParsePlaceholders(pattern, locales, true)
	require.NoError(t, err)

	// Verify results
	assert.Len(t, results, 1)
	assert.Equal(t, "validation", results[0].Kind)
	assert.Len(t, results[0].Items, 2)

	// Verify compound format processing
	emailItem := results[0].Items["EmailAddress"]
	assert.Equal(t, "メールアドレス", emailItem["ja"])
	assert.Equal(t, "Email Address", emailItem["en"])
}

func TestParsePlaceholdersEmptyPattern(t *testing.T) {
	// Test with non-existent pattern (should return empty, not error)
	results, err := ParsePlaceholders("/nonexistent/*.yaml", []string{"ja", "en"}, false)
	assert.NoError(t, err, "Should not return error for non-existent placeholders")
	assert.Empty(t, results, "Should return empty slice for non-existent placeholders")
}

func TestParsePlaceholdersInvalidKindName(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_invalid_kind_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create file with invalid kind name (contains hyphens)
	invalidFile := filepath.Join(tempDir, "invalid-kind.yaml")
	invalidContent := `Item1: "Value1"`
	require.NoError(t, os.WriteFile(invalidFile, []byte(invalidContent), 0644))

	// Execute ParsePlaceholders - should return validation error
	pattern := filepath.Join(tempDir, "*.yaml")
	results, err := ParsePlaceholders(pattern, []string{"ja"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid placeholder kind name")
	assert.Nil(t, results)
}

func TestParsePlaceholdersInvalidItemID(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_invalid_id_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create file with invalid item ID (contains hyphens)
	invalidFile := filepath.Join(tempDir, "valid.yaml")
	invalidContent := `invalid-id: "Value1"`
	require.NoError(t, os.WriteFile(invalidFile, []byte(invalidContent), 0644))

	// Execute ParsePlaceholders - should return validation error
	pattern := filepath.Join(tempDir, "*.yaml")
	results, err := ParsePlaceholders(pattern, []string{"ja"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid placeholder item ID")
	assert.Nil(t, results)
}

func TestIsValidGoIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid simple", "field", true},
		{"valid with underscore", "field_name", true},
		{"valid with numbers", "field1", true},
		{"valid starting with underscore", "_field", true},
		{"valid camelCase", "fieldName", true},
		{"invalid empty", "", false},
		{"invalid starting with number", "1field", false},
		{"invalid with hyphen", "field-name", false},
		{"invalid with space", "field name", false},
		{"invalid with special chars", "field@name", false},
		{"invalid with dot", "field.name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidGoIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectLocale(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"japanese locale", "field.ja.yaml", "ja"},
		{"english locale", "field.en.yaml", "en"},
		{"multiple dots", "field.ja.backup.yaml", "ja"},
		{"no locale", "field.yaml", "yaml"},
		{"single name", "field", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLocale(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeCompoundFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_decode_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test YAML compound file
	yamlFile := filepath.Join(tempDir, "test.yaml")
	yamlContent := `Item1:
  ja: "値1"
  en: "Value1"
Item2:
  ja: "値2"
  en: "Value2"`
	require.NoError(t, os.WriteFile(yamlFile, []byte(yamlContent), 0644))

	f, err := os.Open(yamlFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	result, err := decodeCompoundFile(f, ".yaml")
	require.NoError(t, err)

	expected := map[string]map[string]string{
		"Item1": {"ja": "値1", "en": "Value1"},
		"Item2": {"ja": "値2", "en": "Value2"},
	}
	assert.Equal(t, expected, result)
}

func TestDecodeSimpleFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_decode_simple_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test YAML simple file
	yamlFile := filepath.Join(tempDir, "simple.yaml")
	yamlContent := `Item1: "Value1"
Item2: "Value2"`
	require.NoError(t, os.WriteFile(yamlFile, []byte(yamlContent), 0644))

	f, err := os.Open(yamlFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	result, err := decodeSimpleFile(f, ".yaml")
	require.NoError(t, err)

	expected := map[string]string{
		"Item1": "Value1",
		"Item2": "Value2",
	}
	assert.Equal(t, expected, result)
}

func TestParsePlaceholdersJSONFormat(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_json_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create JSON compound format file
	jsonFile := filepath.Join(tempDir, "field.json")
	jsonContent := `{
  "EmailAddress": {
    "ja": "メールアドレス",
    "en": "Email Address"
  },
  "Required": {
    "ja": "必須",
    "en": "Required"
  }
}`
	require.NoError(t, os.WriteFile(jsonFile, []byte(jsonContent), 0644))

	// Execute ParsePlaceholders with JSON format
	pattern := filepath.Join(tempDir, "*.json")
	results, err := ParsePlaceholders(pattern, []string{"ja", "en"}, true)
	require.NoError(t, err)

	// Verify JSON parsing
	assert.Len(t, results, 1)
	assert.Equal(t, "field", results[0].Kind)
	assert.Len(t, results[0].Items, 2)

	emailItem := results[0].Items["EmailAddress"]
	assert.Equal(t, "メールアドレス", emailItem["ja"])
	assert.Equal(t, "Email Address", emailItem["en"])
}
