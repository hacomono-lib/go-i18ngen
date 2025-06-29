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

func (s *ParserTestSuite) TestParsePlaceholders() {
	// Create test placeholder files
	// Simple format files
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

	// Execute ParsePlaceholders
	pattern := filepath.Join(s.tempDir, "field.*.yaml")
	locales := []string{"ja", "en"}
	results, err := ParsePlaceholders(pattern, locales, false)
	s.Require().NoError(err)

	// Verify results
	s.Len(results, 1, "Should have one placeholder source")
	s.Equal("field", results[0].Kind)
	s.Len(results[0].Items, 3, "Should have three items")

	// Verify specific items
	s.Contains(results[0].Items, "EmailAddress")
	s.Contains(results[0].Items, "FirstName")
	s.Contains(results[0].Items, "LastName")

	// Verify locales
	emailItem := results[0].Items["EmailAddress"]
	s.Equal("メールアドレス", emailItem["ja"])
	s.Equal("Email Address", emailItem["en"])
}

func (s *ParserTestSuite) TestParsePlaceholdersCompoundFormat() {
	// Create compound format file
	compoundFile := filepath.Join(s.tempDir, "validation.yaml")
	compoundContent := `EmailAddress:
  ja: "メールアドレス"
  en: "Email Address"
Required:
  ja: "必須"
  en: "Required"`
	s.Require().NoError(os.WriteFile(compoundFile, []byte(compoundContent), 0644))

	// Execute ParsePlaceholders with compound format
	pattern := filepath.Join(s.tempDir, "validation.yaml")
	locales := []string{"ja", "en"}
	results, err := ParsePlaceholders(pattern, locales, true)
	s.Require().NoError(err)

	// Verify results
	s.Len(results, 1)
	s.Equal("validation", results[0].Kind)
	s.Len(results[0].Items, 2)

	// Verify compound format processing
	emailItem := results[0].Items["EmailAddress"]
	s.Equal("メールアドレス", emailItem["ja"])
	s.Equal("Email Address", emailItem["en"])
}

func (s *ParserTestSuite) TestParsePlaceholdersEmptyPattern() {
	// Test with non-existent pattern (should return empty, not error)
	results, err := ParsePlaceholders("/nonexistent/*.yaml", []string{"ja", "en"}, false)
	s.NoError(err, "Should not return error for non-existent placeholders")
	s.Empty(results, "Should return empty slice for non-existent placeholders")
}

func (s *ParserTestSuite) TestParsePlaceholdersInvalidKindName() {
	// Create file with invalid kind name (contains hyphens)
	invalidFile := filepath.Join(s.tempDir, "invalid-kind.yaml")
	invalidContent := `Item1: "Value1"`
	s.Require().NoError(os.WriteFile(invalidFile, []byte(invalidContent), 0644))

	// Execute ParsePlaceholders - should return validation error
	pattern := filepath.Join(s.tempDir, "invalid-kind.yaml")
	results, err := ParsePlaceholders(pattern, []string{"ja"}, false)
	s.Error(err)
	s.Contains(err.Error(), "invalid placeholder kind name")
	s.Nil(results)
}

func (s *ParserTestSuite) TestParsePlaceholdersInvalidItemID() {
	// Create file with invalid item ID (contains hyphens)
	invalidFile := filepath.Join(s.tempDir, "valid_kind.yaml")
	invalidContent := `invalid-id: "Value1"`
	s.Require().NoError(os.WriteFile(invalidFile, []byte(invalidContent), 0644))

	// Execute ParsePlaceholders - should return validation error
	pattern := filepath.Join(s.tempDir, "valid_kind.yaml")
	results, err := ParsePlaceholders(pattern, []string{"ja"}, false)
	s.Error(err)
	s.Contains(err.Error(), "invalid placeholder item ID")
	s.Nil(results)
}

func (s *ParserTestSuite) TestParsePlaceholdersJSONFormat() {
	// Create JSON compound format file
	jsonFile := filepath.Join(s.tempDir, "field_json.json")
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
	s.Require().NoError(os.WriteFile(jsonFile, []byte(jsonContent), 0644))

	// Execute ParsePlaceholders with JSON format
	pattern := filepath.Join(s.tempDir, "field_json.json")
	results, err := ParsePlaceholders(pattern, []string{"ja", "en"}, true)
	s.Require().NoError(err)

	// Verify JSON parsing
	s.Len(results, 1)
	s.Equal("field_json", results[0].Kind)
	s.Len(results[0].Items, 2)

	emailItem := results[0].Items["EmailAddress"]
	s.Equal("メールアドレス", emailItem["ja"])
	s.Equal("Email Address", emailItem["en"])
}

func (s *ParserTestSuite) TestIsValidGoIdentifier() {
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
		s.Run(tt.name, func() {
			result := isValidGoIdentifier(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ParserTestSuite) TestDetectLocale() {
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
		s.Run(tt.name, func() {
			result := detectLocale(tt.filename)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ParserTestSuite) TestDecodeCompoundFile() {
	// Test YAML compound file
	yamlFile := filepath.Join(s.tempDir, "compound_test.yaml")
	yamlContent := `Item1:
  ja: "値1"
  en: "Value1"
Item2:
  ja: "値2"
  en: "Value2"`
	s.Require().NoError(os.WriteFile(yamlFile, []byte(yamlContent), 0644))

	f, err := os.Open(yamlFile)
	s.Require().NoError(err)
	defer func() { _ = f.Close() }()

	result, err := decodeCompoundFile(f, ".yaml")
	s.Require().NoError(err)

	expected := map[string]map[string]string{
		"Item1": {"ja": "値1", "en": "Value1"},
		"Item2": {"ja": "値2", "en": "Value2"},
	}
	s.Equal(expected, result)
}

func (s *ParserTestSuite) TestDecodeSimpleFile() {
	// Test YAML simple file
	yamlFile := filepath.Join(s.tempDir, "simple_test.yaml")
	yamlContent := `Item1: "Value1"
Item2: "Value2"`
	s.Require().NoError(os.WriteFile(yamlFile, []byte(yamlContent), 0644))

	f, err := os.Open(yamlFile)
	s.Require().NoError(err)
	defer func() { _ = f.Close() }()

	result, err := decodeSimpleFile(f, ".yaml")
	s.Require().NoError(err)

	expected := map[string]string{
		"Item1": "Value1",
		"Item2": "Value2",
	}
	s.Equal(expected, result)
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

// Run the test suite
func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}