package model

import (
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/templatex"

	"github.com/stretchr/testify/suite"
)

type TemplateProcessorTestSuite struct {
	suite.Suite
	testConfig *config.Config
}

func (s *TemplateProcessorTestSuite) SetupSuite() {
	s.testConfig = &config.Config{
		Locales:           []string{"ja", "en"},
		Compound:          true,
		MessagesGlob:      "./messages/*.yaml",
		PlaceholdersGlob:  "./placeholders/*.yaml",
		OutputDir:         "./",
		OutputPackage:     "i18n",
		PluralPlaceholder: "Count",
	}
}

func (s *TemplateProcessorTestSuite) TestProcessTemplateForDuplicates() {
	tests := []struct {
		name     string
		template string
		fields   []string
		expected string
	}{
		{
			name:     "duplicate placeholders",
			template: "{{.name}}さん、{{.name}}さんのアカウント",
			fields:   []string{"name", "name"},
			expected: "{{.name1}}さん、{{.name2}}さんのアカウント",
		},
		{
			name:     "duplicate placeholders with template functions",
			template: "{{.field}}の{{.field | upper}}検証エラー",
			fields:   []string{"field", "field"},
			expected: "{{.field1}}の{{.field2 | upper}}検証エラー",
		},
		{
			name:     "no duplicates",
			template: "{{.entity}}が見つかりません",
			fields:   []string{"entity"},
			expected: "{{.entity}}が見つかりません",
		},
		{
			name:     "mixed case",
			template: "{{.user}}が{{.action}}を{{.user}}に実行",
			fields:   []string{"user", "action", "user"},
			expected: "{{.user1}}が{{.action}}を{{.user2}}に実行",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := processTemplateForDuplicates(tt.template, tt.fields)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *TemplateProcessorTestSuite) TestProcessMessageTemplates() {
	templates := map[string]string{
		"ja": "{{.name}}さん、{{.name}}さんのアカウント",
		"en": "Welcome {{.name}}, to {{.name}}'s account!",
	}
	fields := []string{"name", "name"}

	result := ProcessMessageTemplates(templates, fields)

	expected := map[string]string{
		"ja": "{{.name1}}さん、{{.name2}}さんのアカウント",
		"en": "Welcome {{.name1}}, to {{.name2}}'s account!",
	}

	s.Equal(expected, result)
}

func (s *TemplateProcessorTestSuite) TestProcessMessageTemplatesWithFieldInfos() {
	testCases := []struct {
		name       string
		templates  map[string]string
		fieldInfos []FieldInfo
		expected   map[string]string
	}{
		{
			name: "suffix notation basic",
			templates: map[string]string{
				"en": "From {{.entity:from}} to {{.entity:to}}",
				"ja": "{{.entity:from}}から{{.entity:to}}へ",
			},
			fieldInfos: []FieldInfo{
				{Name: "entity", Suffix: "from"},
				{Name: "entity", Suffix: "to"},
			},
			expected: map[string]string{
				"en": "From {{.entityFrom}} to {{.entityTo}}",
				"ja": "{{.entityFrom}}から{{.entityTo}}へ",
			},
		},
		{
			name: "suffix notation with template functions",
			templates: map[string]string{
				"en": "{{.entity:from | title}} to {{.entity:to | upper}}",
			},
			fieldInfos: []FieldInfo{
				{Name: "entity", Suffix: "from"},
				{Name: "entity", Suffix: "to"},
			},
			expected: map[string]string{
				"en": "{{.entityFrom | title}} to {{.entityTo | upper}}",
			},
		},
		{
			name: "mixed suffix and regular placeholders",
			templates: map[string]string{
				"en": "{{.name}} moved {{.entity:from}} to {{.entity:to}}",
			},
			fieldInfos: []FieldInfo{
				{Name: "name", Suffix: ""},
				{Name: "entity", Suffix: "from"},
				{Name: "entity", Suffix: "to"},
			},
			expected: map[string]string{
				"en": "{{.name}} moved {{.entityFrom}} to {{.entityTo}}",
			},
		},
		{
			name: "numeric and complex suffixes",
			templates: map[string]string{
				"en": "{{.item:1}} and {{.field:input_value}}",
			},
			fieldInfos: []FieldInfo{
				{Name: "item", Suffix: "1"},
				{Name: "field", Suffix: "input_value"},
			},
			expected: map[string]string{
				"en": "{{.item1}} and {{.fieldInputValue}}",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := ProcessMessageTemplatesWithFieldInfos(tc.templates, tc.fieldInfos)
			s.Equal(tc.expected, result)
		})
	}
}

func (s *TemplateProcessorTestSuite) TestBuild() {
	// Create test messages
	messages := []MessageSource{
		{
			ID: "ValidationError",
			Templates: map[string]string{
				"ja": "{{.field}}は{{.reason}}です",
				"en": "{{.field}} is {{.reason}}",
			},
			FieldInfos: []FieldInfo{
				{Name: "field", Suffix: ""},
				{Name: "reason", Suffix: ""},
			},
		},
		{
			ID: "UserWelcome",
			Templates: map[string]string{
				"ja": "{{.name:user}}さん、{{.name:owner}}さんのアカウントです",
				"en": "Welcome {{.name:user}}, to {{.name:owner}}'s account",
			},
			FieldInfos: []FieldInfo{
				{Name: "name", Suffix: "user"},
				{Name: "name", Suffix: "owner"},
			},
		},
	}

	// Create test placeholders
	placeholders := []PlaceholderSource{
		{
			Kind: "Field",
			Items: map[string]map[string]string{
				"EmailAddress": {
					"ja": "メールアドレス",
					"en": "Email Address",
				},
				"Password": {
					"ja": "パスワード",
					"en": "Password",
				},
			},
		},
		{
			Kind: "Validation",
			Items: map[string]map[string]string{
				"Required": {
					"ja": "必須です",
					"en": "is required",
				},
				"Invalid": {
					"ja": "無効です",
					"en": "is invalid",
				},
			},
		},
	}

	// Execute Build
	locales := []string{"ja", "en"}
	result, err := Build(messages, placeholders, locales, s.testConfig)
	s.Require().NoError(err)

	// Verify messages
	s.Len(result.Messages, 2)

	// Find messages by ID (order might vary)
	var validationMsg, userWelcomeMsg *templatex.Message
	for i := range result.Messages {
		switch result.Messages[i].ID {
		case "ValidationError":
			validationMsg = &result.Messages[i]
		case "UserWelcome":
			userWelcomeMsg = &result.Messages[i]
		}
	}
	s.Require().NotNil(validationMsg, "ValidationError message should exist")
	s.Require().NotNil(userWelcomeMsg, "UserWelcome message should exist")

	// Check ValidationError message
	s.Equal("ValidationError", validationMsg.ID)
	s.Equal("ValidationError", validationMsg.StructName)
	s.Len(validationMsg.Fields, 2)

	// Find fields by type/template key (order might vary)
	var fieldField, reasonField *templatex.Field
	for i := range validationMsg.Fields {
		switch validationMsg.Fields[i].TemplateKey {
		case "field":
			fieldField = &validationMsg.Fields[i]
		case "reason":
			reasonField = &validationMsg.Fields[i]
		}
	}
	s.Require().NotNil(fieldField, "field should exist")
	s.Require().NotNil(reasonField, "reason should exist")

	// Check field types
	s.Equal("Field", fieldField.FieldName)
	s.Equal("FieldValue", fieldField.Type) // Auto-generated since "field" is not in placeholders
	s.Equal("field", fieldField.TemplateKey)

	s.Equal("Reason", reasonField.FieldName)
	s.Equal("ReasonValue", reasonField.Type) // Auto-generated Value type
	s.Equal("reason", reasonField.TemplateKey)

	// Check UserWelcome message with suffix notation
	s.Equal("UserWelcome", userWelcomeMsg.ID)
	s.Len(userWelcomeMsg.Fields, 2)

	// Find suffix-based fields (order might vary)
	var nameUserField, nameOwnerField *templatex.Field
	for i := range userWelcomeMsg.Fields {
		switch userWelcomeMsg.Fields[i].TemplateKey {
		case "nameUser":
			nameUserField = &userWelcomeMsg.Fields[i]
		case "nameOwner":
			nameOwnerField = &userWelcomeMsg.Fields[i]
		}
	}
	s.Require().NotNil(nameUserField, "nameUser field should exist")
	s.Require().NotNil(nameOwnerField, "nameOwner field should exist")

	// Check suffix-based field names
	s.Equal("NameUser", nameUserField.FieldName)
	s.Equal("NameValue", nameUserField.Type)
	s.Equal("nameUser", nameUserField.TemplateKey)

	s.Equal("NameOwner", nameOwnerField.FieldName)
	s.Equal("NameValue", nameOwnerField.Type)
	s.Equal("nameOwner", nameOwnerField.TemplateKey)

	// Verify placeholders (total count includes auto-generated ones)
	s.GreaterOrEqual(len(result.Placeholders), 2, "Should have at least original placeholders")

	// Check that auto-generated placeholder exists
	var reasonPlaceholder *templatex.Placeholder
	for i := range result.Placeholders {
		if result.Placeholders[i].StructName == "ReasonValue" {
			reasonPlaceholder = &result.Placeholders[i]
			break
		}
	}
	s.Require().NotNil(reasonPlaceholder, "Auto-generated ReasonValue placeholder should exist")
	s.True(reasonPlaceholder.IsValue)
}

func (s *TemplateProcessorTestSuite) TestBuildWithMessageStartingWithDigit() {
	// Test message ID starting with digit (should be prefixed with "Msg")
	messages := []MessageSource{
		{
			ID: "404Error",
			Templates: map[string]string{
				"ja": "{{.message}}",
				"en": "{{.message}}",
			},
			FieldInfos: []FieldInfo{
				{Name: "message", Suffix: ""},
			},
		},
	}

	result, err := Build(messages, []PlaceholderSource{}, []string{"ja", "en"}, s.testConfig)
	s.Require().NoError(err)

	// Verify struct name is prefixed with "Msg"
	s.Equal("Msg404Error", result.Messages[0].StructName)
}

func (s *TemplateProcessorTestSuite) TestBuildTemplates() {
	// Create test data
	messages := []MessageSource{
		{
			ID: "TestMessage",
			Templates: map[string]string{
				"ja": "{{.field}}のテスト",
				"en": "Test for {{.field}}",
			},
			FieldInfos: []FieldInfo{
				{Name: "field", Suffix: ""},
			},
		},
	}

	placeholders := []PlaceholderSource{
		{
			Kind: "Field",
			Items: map[string]map[string]string{
				"EmailAddress": {
					"ja": "メールアドレス",
					"en": "Email Address",
				},
			},
		},
	}

	// Execute BuildTemplates
	locales := []string{"ja", "en"}
	messageTemplates, placeholderTemplates, err := BuildTemplates(messages, placeholders, locales)
	s.Require().NoError(err)

	// Verify message templates
	s.Len(messageTemplates, 1)
	msgTemplate := messageTemplates[0]
	s.Equal("TestMessage", msgTemplate.ID)
	s.Equal("{{.field}}のテスト", msgTemplate.Templates["ja"])
	s.Equal("Test for {{.field}}", msgTemplate.Templates["en"])

	// Verify placeholder templates
	s.Len(placeholderTemplates, 1)
	phTemplate := placeholderTemplates[0]
	s.Equal("Field", phTemplate.Name)
	s.True(phTemplate.HasLocaleFiles)
	s.Contains(phTemplate.LocaleTemplates, "EmailAddress")

	emailTemplates := phTemplate.LocaleTemplates["EmailAddress"]
	s.Equal("メールアドレス", emailTemplates["ja"])
	s.Equal("Email Address", emailTemplates["en"])
}

func (s *TemplateProcessorTestSuite) TestGenerateStructName() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal ID", "ValidationError", "ValidationError"},
		{"camelCase ID", "userWelcome", "UserWelcome"},
		{"snake_case ID", "user_welcome", "UserWelcome"},
		{"ID starting with digit", "404Error", "Msg404Error"},
		{"ID with numbers", "error404", "Error404"},
		{"single word", "error", "Error"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := generateStructName(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *TemplateProcessorTestSuite) TestFieldInfoString() {
	tests := []struct {
		name     string
		field    FieldInfo
		expected string
	}{
		{"field without suffix", FieldInfo{Name: "name", Suffix: ""}, "name"},
		{"field with suffix", FieldInfo{Name: "name", Suffix: "user"}, "name:user"},
		{"field with numeric suffix", FieldInfo{Name: "value", Suffix: "1"}, "value:1"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := tt.field.String()
			s.Equal(tt.expected, result)
		})
	}
}

func (s *TemplateProcessorTestSuite) TestFieldInfoGenerateFieldName() {
	tests := []struct {
		name     string
		field    FieldInfo
		expected string
	}{
		{"field without suffix", FieldInfo{Name: "email_address", Suffix: ""}, "EmailAddress"},
		{"field with suffix", FieldInfo{Name: "name", Suffix: "user"}, "NameUser"},
		{"field with complex suffix", FieldInfo{Name: "entity", Suffix: "from_location"}, "EntityFromLocation"},
		{"field with numeric suffix", FieldInfo{Name: "value", Suffix: "1"}, "Value1"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := tt.field.GenerateFieldName()
			s.Equal(tt.expected, result)
		})
	}
}

func (s *TemplateProcessorTestSuite) TestFieldInfoGenerateTemplateKey() {
	tests := []struct {
		name     string
		field    FieldInfo
		expected string
	}{
		{"basic field", FieldInfo{Name: "username", Suffix: ""}, "username"},
		{"field with simple suffix", FieldInfo{Name: "title", Suffix: "admin"}, "titleAdmin"},
		{"field with underscore suffix", FieldInfo{Name: "location", Suffix: "from_state"}, "locationFromState"},
		{"field with number suffix", FieldInfo{Name: "counter", Suffix: "3"}, "counter3"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := tt.field.GenerateTemplateKey()
			s.Equal(tt.expected, result)
		})
	}
}

func (s *TemplateProcessorTestSuite) TestBuildEmptyInput() {
	// Test with empty messages and placeholders
	result, err := Build([]MessageSource{}, []PlaceholderSource{}, []string{"ja", "en"}, s.testConfig)
	s.Require().NoError(err)

	s.Empty(result.Messages)
	s.Empty(result.Placeholders)
}

func (s *TemplateProcessorTestSuite) TestBuildWithEmptyLocales() {
	// Test with empty locales
	messages := []MessageSource{
		{
			ID: "TestMessage",
			Templates: map[string]string{
				"ja": "テスト",
			},
			FieldInfos: []FieldInfo{},
		},
	}

	result, err := Build(messages, []PlaceholderSource{}, []string{}, s.testConfig)
	s.Require().NoError(err)

	s.Len(result.Messages, 1)
	s.Equal("TestMessage", result.Messages[0].ID)
}

// Run the test suite
func (s *TemplateProcessorTestSuite) TestExtractTemplateFunctionsBasic() {
	// Test the basic extractTemplateFunctions helper function
	tests := []struct {
		name             string
		templateFunction string
		expected         []string
	}{
		{
			"no functions",
			"",
			nil,
		},
		{
			"single function",
			"| title",
			[]string{"title"},
		},
		{
			"multiple functions",
			"| title | upper",
			[]string{"title", "upper"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := extractTemplateFunctions(tt.templateFunction)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *TemplateProcessorTestSuite) TestBuildTemplateFunctionsMetadata() {
	messages := []MessageSource{
		{
			ID: "ValidationError",
			Templates: map[string]string{
				"en": "{{.field | title}} is {{.status | upper}}",
				"ja": "{{.field}}は{{.status}}です",
			},
			FieldInfos: []FieldInfo{
				{Name: "field", Suffix: ""},
				{Name: "status", Suffix: ""},
			},
		},
		{
			ID: "TransferMessage",
			Templates: map[string]string{
				"en": "From {{.entity:from | title}} to {{.entity:to}}",
				"ja": "{{.entity:from}}から{{.entity:to}}へ",
			},
			FieldInfos: []FieldInfo{
				{Name: "entity", Suffix: "from"},
				{Name: "entity", Suffix: "to"},
			},
		},
	}

	result := BuildTemplateFunctionsMetadata(messages, []string{"en", "ja"})

	expected := map[string]map[string]map[string][]string{
		"ValidationError": {
			"en": {
				"field":  {"title"},
				"status": {"upper"},
			},
		},
		"TransferMessage": {
			"en": {
				"entityFrom": {"title"},
			},
		},
	}

	s.Equal(expected, result)
}

func (s *TemplateProcessorTestSuite) TestExtractTemplateFunctionsFromTemplate() {
	tests := []struct {
		name      string
		template  string
		fieldInfo FieldInfo
		expected  []string
	}{
		{
			"no functions",
			"Simple {{.field}} message",
			FieldInfo{Name: "field", Suffix: ""},
			nil,
		},
		{
			"single function",
			"{{.field | title}} message",
			FieldInfo{Name: "field", Suffix: ""},
			[]string{"title"},
		},
		{
			"suffix notation with function",
			"From {{.entity:from | title}} to destination",
			FieldInfo{Name: "entity", Suffix: "from"},
			[]string{"title"},
		},
		{
			"chained functions",
			"{{.field | title | lower}} is processed",
			FieldInfo{Name: "field", Suffix: ""},
			[]string{"title", "lower"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := extractTemplateFunctionsFromTemplate(tt.template, tt.fieldInfo)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *TemplateProcessorTestSuite) TestExtractTemplateFunctionsEdgeCases() {
	tests := []struct {
		name             string
		templateFunction string
		expected         []string
	}{
		{
			"empty string",
			"",
			nil,
		},
		{
			"single function",
			"| title",
			[]string{"title"},
		},
		{
			"multiple functions",
			"| title | upper | lower",
			[]string{"title", "upper", "lower"},
		},
		{
			"whitespace handling",
			"|  title  |  upper  ",
			[]string{"title", "upper"},
		},
		{
			"no leading pipe",
			"title | upper",
			[]string{"title", "upper"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := extractTemplateFunctions(tt.templateFunction)
			s.Equal(tt.expected, result)
		})
	}
}

func TestTemplateProcessorSuite(t *testing.T) {
	suite.Run(t, new(TemplateProcessorTestSuite))
}
