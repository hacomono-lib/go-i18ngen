package model

import (
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/templatex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessTemplateForDuplicates(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := processTemplateForDuplicates(tt.template, tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessMessageTemplates(t *testing.T) {
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

	assert.Equal(t, expected, result)
}

func TestProcessMessageTemplatesWithFieldInfos(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			result := ProcessMessageTemplatesWithFieldInfos(tc.templates, tc.fieldInfos)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Tests for Build function
func TestBuild(t *testing.T) {
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
	result, err := Build(messages, placeholders, locales)
	require.NoError(t, err)

	// Verify messages
	assert.Len(t, result.Messages, 2)

	// Find messages by ID (order might vary)
	var validationMsg, userWelcomeMsg *templatex.Message
	for i := range result.Messages {
		if result.Messages[i].ID == "ValidationError" {
			validationMsg = &result.Messages[i]
		} else if result.Messages[i].ID == "UserWelcome" {
			userWelcomeMsg = &result.Messages[i]
		}
	}
	require.NotNil(t, validationMsg, "ValidationError message should exist")
	require.NotNil(t, userWelcomeMsg, "UserWelcome message should exist")

	// Check ValidationError message
	assert.Equal(t, "ValidationError", validationMsg.ID)
	assert.Equal(t, "ValidationError", validationMsg.StructName)
	assert.Len(t, validationMsg.Fields, 2)

	// Find fields by type/template key (order might vary)
	var fieldField, reasonField *templatex.Field
	for i := range validationMsg.Fields {
		if validationMsg.Fields[i].TemplateKey == "field" {
			fieldField = &validationMsg.Fields[i]
		} else if validationMsg.Fields[i].TemplateKey == "reason" {
			reasonField = &validationMsg.Fields[i]
		}
	}
	require.NotNil(t, fieldField, "field should exist")
	require.NotNil(t, reasonField, "reason should exist")

	// Check field types
	assert.Equal(t, "Field", fieldField.FieldName)
	assert.Equal(t, "FieldValue", fieldField.Type) // Auto-generated since "field" is not in placeholders
	assert.Equal(t, "field", fieldField.TemplateKey)

	assert.Equal(t, "Reason", reasonField.FieldName)
	assert.Equal(t, "ReasonValue", reasonField.Type) // Auto-generated Value type
	assert.Equal(t, "reason", reasonField.TemplateKey)

	// Check UserWelcome message with suffix notation
	assert.Equal(t, "UserWelcome", userWelcomeMsg.ID)
	assert.Len(t, userWelcomeMsg.Fields, 2)

	// Find suffix-based fields (order might vary)
	var nameUserField, nameOwnerField *templatex.Field
	for i := range userWelcomeMsg.Fields {
		if userWelcomeMsg.Fields[i].TemplateKey == "nameUser" {
			nameUserField = &userWelcomeMsg.Fields[i]
		} else if userWelcomeMsg.Fields[i].TemplateKey == "nameOwner" {
			nameOwnerField = &userWelcomeMsg.Fields[i]
		}
	}
	require.NotNil(t, nameUserField, "nameUser field should exist")
	require.NotNil(t, nameOwnerField, "nameOwner field should exist")

	// Check suffix-based field names
	assert.Equal(t, "NameUser", nameUserField.FieldName)
	assert.Equal(t, "NameValue", nameUserField.Type)
	assert.Equal(t, "nameUser", nameUserField.TemplateKey)

	assert.Equal(t, "NameOwner", nameOwnerField.FieldName)
	assert.Equal(t, "NameValue", nameOwnerField.Type)
	assert.Equal(t, "nameOwner", nameOwnerField.TemplateKey)

	// Verify placeholders (total count includes auto-generated ones)
	assert.GreaterOrEqual(t, len(result.Placeholders), 2, "Should have at least original placeholders")

	// Check that auto-generated placeholder exists
	var reasonPlaceholder *templatex.Placeholder
	for i := range result.Placeholders {
		if result.Placeholders[i].StructName == "ReasonValue" {
			reasonPlaceholder = &result.Placeholders[i]
			break
		}
	}
	require.NotNil(t, reasonPlaceholder, "Auto-generated ReasonValue placeholder should exist")
	assert.True(t, reasonPlaceholder.IsValue)
}

func TestBuildWithMessageStartingWithDigit(t *testing.T) {
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

	result, err := Build(messages, []PlaceholderSource{}, []string{"ja", "en"})
	require.NoError(t, err)

	// Verify struct name is prefixed with "Msg"
	assert.Equal(t, "Msg404Error", result.Messages[0].StructName)
}

// Tests for BuildTemplates function
func TestBuildTemplates(t *testing.T) {
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
	require.NoError(t, err)

	// Verify message templates
	assert.Len(t, messageTemplates, 1)
	msgTemplate := messageTemplates[0]
	assert.Equal(t, "TestMessage", msgTemplate.ID)
	assert.Equal(t, "{{.field}}のテスト", msgTemplate.Templates["ja"])
	assert.Equal(t, "Test for {{.field}}", msgTemplate.Templates["en"])

	// Verify placeholder templates
	assert.Len(t, placeholderTemplates, 1)
	phTemplate := placeholderTemplates[0]
	assert.Equal(t, "Field", phTemplate.Name)
	assert.True(t, phTemplate.HasLocaleFiles)
	assert.Contains(t, phTemplate.LocaleTemplates, "EmailAddress")

	emailTemplates := phTemplate.LocaleTemplates["EmailAddress"]
	assert.Equal(t, "メールアドレス", emailTemplates["ja"])
	assert.Equal(t, "Email Address", emailTemplates["en"])
}

// Tests for generateStructName function
func TestGenerateStructName(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := generateStructName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for FieldInfo methods
func TestFieldInfoString(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFieldInfoGenerateFieldName(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.GenerateFieldName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFieldInfoGenerateTemplateKey(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.GenerateTemplateKey()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test edge cases and error conditions
func TestBuildEmptyInput(t *testing.T) {
	// Test with empty messages and placeholders
	result, err := Build([]MessageSource{}, []PlaceholderSource{}, []string{"ja", "en"})
	require.NoError(t, err)

	assert.Empty(t, result.Messages)
	assert.Empty(t, result.Placeholders)
}

func TestBuildWithEmptyLocales(t *testing.T) {
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

	result, err := Build(messages, []PlaceholderSource{}, []string{})
	require.NoError(t, err)

	assert.Len(t, result.Messages, 1)
	assert.Equal(t, "TestMessage", result.Messages[0].ID)
}
