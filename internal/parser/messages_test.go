package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/model"

	"github.com/stretchr/testify/suite"
)

type ParserTestSuite struct {
	suite.Suite
	tempDir string
}

func (s *ParserTestSuite) SetupSuite() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "i18ngen_parser_test")
	s.Require().NoError(err)
}

func (s *ParserTestSuite) TearDownSuite() {
	if s.tempDir != "" {
		_ = os.RemoveAll(s.tempDir)
	}
}

func (s *ParserTestSuite) TestExtractFieldInfos() {
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
		s.Run(tt.name, func() {
			result := extractFieldInfos(tt.template)
			s.Equal(tt.expected, result, "Field extraction does not match expected values")
		})
	}
}

func (s *ParserTestSuite) TestParseMessages() {
	// Create test message file with only valid syntax (no duplicate placeholders)
	messageFile := filepath.Join(s.tempDir, "messages.yaml")
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
	s.Require().NoError(os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Execute ParseMessages
	pattern := filepath.Join(s.tempDir, "messages.yaml")
	results, err := ParseMessages(pattern)
	s.Require().NoError(err)

	// Verify results
	s.Len(results, 3, "Number of messages does not match expected")

	// Verify EntityNotFound
	entityNotFound := s.findMessageByID(results, "EntityNotFound")
	s.Require().NotNil(entityNotFound, "EntityNotFound message not found")

	expectedEntityFields := []model.FieldInfo{{Name: "entity", Suffix: ""}, {Name: "reason", Suffix: ""}}
	s.Equal(expectedEntityFields, entityNotFound.FieldInfos)
	s.Equal("{{.entity}}が見つかりません: {{.reason}}", entityNotFound.Templates["ja"])
	s.Equal("{{.entity}} not found: {{.reason}}", entityNotFound.Templates["en"])

	// Verify SuffixExample (suffix notation)
	suffixExample := s.findMessageByID(results, "SuffixExample")
	s.Require().NotNil(suffixExample, "SuffixExample message not found")

	expectedSuffixFields := []model.FieldInfo{{Name: "name", Suffix: "user"}, {Name: "name", Suffix: "owner"}}
	s.Equal(expectedSuffixFields, suffixExample.FieldInfos, "Suffix notation placeholders are not properly processed")

	// Verify TemplateFunctionExample
	templateFunctionExample := s.findMessageByID(results, "TemplateFunctionExample")
	s.Require().NotNil(templateFunctionExample, "TemplateFunctionExample message not found")

	expectedTemplateFields := []model.FieldInfo{{Name: "field", Suffix: "input"}, {Name: "field", Suffix: "display"}}
	s.Equal(expectedTemplateFields, templateFunctionExample.FieldInfos, "Placeholders with template functions are not properly processed")
}

func (s *ParserTestSuite) TestParseMessagesWithJSON() {
	// Create JSON format test message file with suffix notation
	messageFile := filepath.Join(s.tempDir, "messages.json")
	messageContent := `{
  "ValidationError": {
    "ja": "{{.field:input}}の{{.field:display | upper}}検証エラー",
    "en": "{{.field:input | title}} validation error for {{.field:display}}"
  }
}`
	s.Require().NoError(os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Execute ParseMessages
	pattern := filepath.Join(s.tempDir, "messages.json")
	results, err := ParseMessages(pattern)
	s.Require().NoError(err)

	// Verify results
	s.Len(results, 1)
	validationError := results[0]
	s.Equal("ValidationError", validationError.ID)

	expectedFields := []model.FieldInfo{{Name: "field", Suffix: "input"}, {Name: "field", Suffix: "display"}}
	s.Equal(expectedFields, validationError.FieldInfos, "Verify that suffix notation and template function processing work with JSON format")
}

func (s *ParserTestSuite) TestParseMessagesDuplicatePlaceholderValidation() {
	// Create test message file with duplicate placeholders (should fail)
	messageFile := filepath.Join(s.tempDir, "invalid_messages.yaml")
	messageContent := `InvalidMessage:
  ja: "{{.name}}さん、{{.name}}さんのアカウントへようこそ"
  en: "Welcome {{.name}}, to {{.name}}'s account"
`
	s.Require().NoError(os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Execute ParseMessages - should return error
	pattern := filepath.Join(s.tempDir, "invalid_messages.yaml")
	results, err := ParseMessages(pattern)
	s.Error(err, "Should return error for duplicate placeholders")
	s.Contains(err.Error(), "duplicate placeholder", "Error message should mention duplicate placeholder")
	s.Contains(err.Error(), "suffix notation", "Error message should suggest suffix notation")
	s.Nil(results)
}

func (s *ParserTestSuite) TestParseMessagesEmptyPattern() {
	// Test with non-existent pattern
	results, err := ParseMessages("/nonexistent/*.yaml")
	s.Error(err, "Should return error for non-existent patterns")
	s.Contains(err.Error(), "no message files found", "Error should indicate no files found")
	s.Nil(results)
}

func (s *ParserTestSuite) TestDecodeMessageFileErrors() {
	// Create invalid YAML file
	invalidFile := filepath.Join(s.tempDir, "invalid.yaml")
	invalidContent := `invalid: yaml: content:
  - unclosed
    brackets: [`
	s.Require().NoError(os.WriteFile(invalidFile, []byte(invalidContent), 0644))

	// Verify that error is returned
	pattern := filepath.Join(s.tempDir, "invalid.yaml")
	results, err := ParseMessages(pattern)
	s.Error(err, "Verify that error is returned for invalid YAML files")
	s.Nil(results)
}

func (s *ParserTestSuite) TestParsePlaceholdersSimpleFormat() {
	// Create test placeholder files in simple format
	fieldFile := filepath.Join(s.tempDir, "field.ja.yaml")
	fieldContent := `EmailAddress: "メールアドレス"
FirstName: "名前"
LastName: "苗字"`
	s.Require().NoError(os.WriteFile(fieldFile, []byte(fieldContent), 0644))

	fieldEnFile := filepath.Join(s.tempDir, "field.en.yaml")
	fieldEnContent := `EmailAddress: "Email Address"
FirstName: "First Name"
LastName: "Last Name"`
	s.Require().NoError(os.WriteFile(fieldEnFile, []byte(fieldEnContent), 0644))

	// Execute ParsePlaceholders for simple format
	pattern := filepath.Join(s.tempDir, "field.*.yaml")
	locales := []string{"ja", "en"}
	results, err := ParsePlaceholders(pattern, locales, false)
	s.Require().NoError(err)

	// Verify results
	s.Len(results, 1, "Should have one placeholder source")
	s.Equal("field", results[0].Kind)
	s.Len(results[0].Items, 3, "Should have three items")

	// Verify content
	s.Equal("メールアドレス", results[0].Items["EmailAddress"]["ja"])
	s.Equal("Email Address", results[0].Items["EmailAddress"]["en"])
}

func (s *ParserTestSuite) TestParsePlaceholdersCompoundFormat() {
	// Create test placeholder files in compound format
	entityFile := filepath.Join(s.tempDir, "entity.yaml")
	entityContent := `user:
  ja: "ユーザー"
  en: "User"
product:
  ja: "製品"
  en: "Product"`
	s.Require().NoError(os.WriteFile(entityFile, []byte(entityContent), 0644))

	// Execute ParsePlaceholders for compound format
	pattern := filepath.Join(s.tempDir, "entity.yaml")
	locales := []string{"ja", "en"}
	results, err := ParsePlaceholders(pattern, locales, true)
	s.Require().NoError(err)

	// Verify results
	s.Len(results, 1, "Should have one placeholder source")
	s.Equal("entity", results[0].Kind)
	s.Len(results[0].Items, 2, "Should have two items")

	// Verify content
	s.Equal("ユーザー", results[0].Items["user"]["ja"])
	s.Equal("User", results[0].Items["user"]["en"])
}

func (s *ParserTestSuite) TestParsePlaceholdersErrorCases() {
	tests := []struct {
		name        string
		setupFunc   func() string // Returns pattern
		expectError bool
	}{
		{
			"non-existent pattern",
			func() string {
				return "/nonexistent/path/*.yaml"
			},
			false, // ParsePlaceholders returns empty slice for missing files
		},
		{
			"invalid YAML content",
			func() string {
				invalidFile := filepath.Join(s.tempDir, "invalid.yaml")
				invalidContent := `invalid: yaml: content:`
				s.Require().NoError(os.WriteFile(invalidFile, []byte(invalidContent), 0644))
				return invalidFile
			},
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			pattern := tt.setupFunc()
			results, err := ParsePlaceholders(pattern, []string{"en"}, true)

			if tt.expectError {
				s.Error(err)
				s.Nil(results)
			} else {
				s.NoError(err)
				s.NotNil(results)
			}
		})
	}
}

// Helper function
func (s *ParserTestSuite) findMessageByID(messages []model.MessageSource, id string) *model.MessageSource {
	for i := range messages {
		if messages[i].ID == id {
			return &messages[i]
		}
	}
	return nil
}

func (s *ParserTestSuite) TestConvertMixedToStringMap() {
	// Test for convertMixedToStringMap function coverage
	tests := []struct {
		name     string
		input    map[string]map[string]interface{}
		expected map[string]map[string]string
	}{
		{
			name: "simple string templates",
			input: map[string]map[string]interface{}{
				"SimpleMessage": {
					"en": "Hello {{.name}}",
					"ja": "こんにちは {{.name}}",
				},
			},
			expected: map[string]map[string]string{
				"SimpleMessage": {
					"en": "Hello {{.name}}",
					"ja": "こんにちは {{.name}}",
				},
			},
		},
		{
			name: "plural templates",
			input: map[string]map[string]interface{}{
				"CountMessage": {
					"en": map[string]interface{}{
						"one":   "{{.Count}} item",
						"other": "{{.Count}} items",
					},
					"ja": "{{.Count}}個のアイテム",
				},
			},
			expected: map[string]map[string]string{
				"CountMessage": {
					"en": "{{.Count}} items", // Should use "other" form
					"ja": "{{.Count}}個のアイテム",
				},
			},
		},
		{
			name: "mixed interface types",
			input: map[string]map[string]interface{}{
				"MixedMessage": {
					"en": map[interface{}]interface{}{
						"one":   "One item",
						"other": "{{.Count}} items",
					},
					"ja": 123, // Non-string fallback
				},
			},
			expected: map[string]map[string]string{
				"MixedMessage": {
					"en": "{{.Count}} items",
					"ja": "123",
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := convertMixedToStringMap(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ParserTestSuite) TestConvertPluralToTemplate() {
	// Test for convertPluralToTemplate function coverage
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected string
	}{
		{
			name: "has other form",
			input: map[string]interface{}{
				"one":   "{{.Count}} item",
				"other": "{{.Count}} items",
			},
			expected: "{{.Count}} items",
		},
		{
			name: "has one form only",
			input: map[string]interface{}{
				"one": "{{.Count}} item",
			},
			expected: "{{.Count}} item",
		},
		{
			name: "has few form only",
			input: map[string]interface{}{
				"few": "{{.Count}} few items",
			},
			expected: "{{.Count}} few items",
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: "{{.Count}} items", // fallback
		},
		{
			name: "non-string values",
			input: map[string]interface{}{
				"other": 123,
			},
			expected: "{{.Count}} items", // fallback when conversion fails
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := convertPluralToTemplate(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ParserTestSuite) TestIsValidGoIdentifier() {
	// Test for isValidGoIdentifier function coverage
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid identifier",
			input:    "validName",
			expected: true,
		},
		{
			name:     "valid identifier with underscore",
			input:    "valid_name",
			expected: true,
		},
		{
			name:     "valid identifier starting with underscore",
			input:    "_validName",
			expected: true,
		},
		{
			name:     "invalid - empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid - starts with number",
			input:    "123invalid",
			expected: false,
		},
		{
			name:     "invalid - contains space",
			input:    "invalid name",
			expected: false,
		},
		{
			name:     "invalid - contains hyphen",
			input:    "invalid-name",
			expected: false,
		},
		{
			name:     "invalid - contains special characters",
			input:    "invalid@name",
			expected: false,
		},
		{
			name:     "valid with numbers",
			input:    "valid123",
			expected: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := isValidGoIdentifier(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ParserTestSuite) TestDetectLocaleEdgeCases() {
	// Test for detectLocale function coverage
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "normal filename",
			filename: "field.en.yaml",
			expected: "en",
		},
		{
			name:     "filename with multiple dots",
			filename: "field.en.test.yaml",
			expected: "en",
		},
		{
			name:     "filename with no locale part",
			filename: "field.yaml",
			expected: "yaml",
		},
		{
			name:     "filename with single part",
			filename: "field",
			expected: "unknown",
		},
		{
			name:     "empty filename",
			filename: "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := detectLocale(tt.filename)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ParserTestSuite) TestDecodeFileErrors() {
	// Test error cases for decode functions
	tempFile := filepath.Join(s.tempDir, "invalid.json")
	invalidJSONContent := `{"invalid": "json", "unclosed": [`
	s.Require().NoError(os.WriteFile(tempFile, []byte(invalidJSONContent), 0644))

	file, err := os.Open(tempFile)
	s.Require().NoError(err)
	defer func() { _ = file.Close() }()

	// Test decodeCompoundFile with invalid JSON
	_, err = decodeCompoundFile(file, ".json")
	s.Error(err, "Should error on invalid JSON")

	// Reset file pointer
	_, _ = file.Seek(0, 0)

	// Test decodeSimpleFile with invalid JSON
	_, err = decodeSimpleFile(file, ".json")
	s.Error(err, "Should error on invalid JSON")
}

func (s *ParserTestSuite) TestDecodeValidFiles() {
	// Test valid file decoding
	// Create valid compound JSON file
	compoundFile := filepath.Join(s.tempDir, "compound.json")
	compoundContent := `{
		"item1": {"en": "Item 1", "ja": "アイテム1"},
		"item2": {"en": "Item 2", "ja": "アイテム2"}
	}`
	s.Require().NoError(os.WriteFile(compoundFile, []byte(compoundContent), 0644))

	file, err := os.Open(compoundFile)
	s.Require().NoError(err)
	defer func() { _ = file.Close() }()

	result, err := decodeCompoundFile(file, ".json")
	s.NoError(err)
	s.Equal("Item 1", result["item1"]["en"])
	s.Equal("アイテム1", result["item1"]["ja"])

	// Create valid simple YAML file
	simpleFile := filepath.Join(s.tempDir, "simple.yaml")
	simpleContent := `item1: "Simple Item 1"
item2: "Simple Item 2"`
	s.Require().NoError(os.WriteFile(simpleFile, []byte(simpleContent), 0644))

	file2, err := os.Open(simpleFile)
	s.Require().NoError(err)
	defer func() { _ = file2.Close() }()

	result2, err := decodeSimpleFile(file2, ".yaml")
	s.NoError(err)
	s.Equal("Simple Item 1", result2["item1"])
	s.Equal("Simple Item 2", result2["item2"])
}

// Run the test suite
func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}
