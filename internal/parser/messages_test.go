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

// Run the test suite
func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}
