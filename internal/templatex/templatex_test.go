package templatex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TemplatexTestSuite struct {
	suite.Suite
	tempDir string
}

func TestTemplatexSuite(t *testing.T) {
	suite.Run(t, new(TemplatexTestSuite))
}

func (s *TemplatexTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "templatex_test")
	s.Require().NoError(err)
}

func (s *TemplatexTestSuite) TearDownTest() {
	_ = os.RemoveAll(s.tempDir)
}

func (s *TemplatexTestSuite) TestRenderGoI18n_Success() {
	outputFile := filepath.Join(s.tempDir, "test.go")

	messages := []MessageTemplate{
		{
			ID: "UserWelcome",
			Templates: map[string]string{
				"ja": "{{.name}}さん、ようこそ！",
				"en": "Welcome, {{.name}}!",
			},
		},
	}

	placeholderTemplates := []PlaceholderTemplate{
		{
			Name:           "name",
			HasLocaleFiles: true,
			LocaleTemplates: map[string]map[string]string{
				"ja": {"name": "名前"},
				"en": {"name": "Name"},
			},
		},
	}

	placeholders := []Placeholder{
		{
			StructName: "NameText",
			VarName:    "nameTemplates",
			IsValue:    false,
			Items: []PlaceholderItem{
				{ID: "user", FieldName: "User", Templates: map[string]string{"ja": "ユーザー", "en": "User"}},
			},
		},
	}

	messageDefs := []Message{
		{
			ID:         "UserWelcome",
			StructName: "UserWelcome",
			Fields: []Field{
				{FieldName: "Name", Type: "NameValue", TemplateKey: "name"},
			},
			Templates: map[string]string{
				"ja": "{{.name}}さん、ようこそ！",
				"en": "Welcome, {{.name}}!",
			},
		},
	}

	err := RenderGoI18n(
		outputFile,
		"testpkg",
		"ja",
		messages,
		placeholderTemplates,
		placeholders,
		messageDefs,
		[]string{"ja", "en"},
	)

	s.Assert().NoError(err)
	s.Assert().FileExists(outputFile)

	// Verify generated content
	content, err := os.ReadFile(outputFile)
	s.Require().NoError(err)

	contentStr := string(content)
	s.Assert().Contains(contentStr, "package testpkg")
	s.Assert().Contains(contentStr, "UserWelcome")
	s.Assert().Contains(contentStr, "NewUserWelcome")
}

func (s *TemplatexTestSuite) TestRenderGoI18n_InvalidOutputPath() {
	// Use an invalid path that cannot be created
	invalidPath := filepath.Join("/invalid", "path", "that", "does", "not", "exist", "test.go")

	err := RenderGoI18n(
		invalidPath,
		"testpkg",
		"ja",
		[]MessageTemplate{},
		[]PlaceholderTemplate{},
		[]Placeholder{},
		[]Message{},
		[]string{"ja"},
	)

	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "failed to write generated code")
}

func (s *TemplatexTestSuite) TestRenderGoI18n_EmptyData() {
	outputFile := filepath.Join(s.tempDir, "empty.go")

	err := RenderGoI18n(
		outputFile,
		"emptypkg",
		"en",
		[]MessageTemplate{},
		[]PlaceholderTemplate{},
		[]Placeholder{},
		[]Message{},
		[]string{"en"},
	)

	s.Assert().NoError(err)
	s.Assert().FileExists(outputFile)

	// Verify generated content
	content, err := os.ReadFile(outputFile)
	s.Require().NoError(err)

	contentStr := string(content)
	s.Assert().Contains(contentStr, "package emptypkg")
}

func (s *TemplatexTestSuite) TestRenderGoI18nWithTemplateFunctions_Success() {
	outputFile := filepath.Join(s.tempDir, "config_test.go")

	templateFunctions := map[string]map[string]map[string][]string{}

	err := RenderGoI18nWithTemplateFunctions(
		outputFile,
		"configpkg",
		"en",
		[]MessageTemplate{},
		[]PlaceholderTemplate{},
		[]Placeholder{},
		[]Message{},
		[]string{"en"},
		templateFunctions,
	)

	s.Assert().NoError(err)
	s.Assert().FileExists(outputFile)

	content, err := os.ReadFile(outputFile)
	s.Require().NoError(err)

	contentStr := string(content)
	s.Assert().Contains(contentStr, "package configpkg")
}

func (s *TemplatexTestSuite) TestRenderGoI18n_WithMinimalData() {
	outputFile := filepath.Join(s.tempDir, "minimal.go")

	// Test with minimal valid data
	messages := []MessageTemplate{
		{
			ID: "Simple",
			Templates: map[string]string{
				"en": "Hello World",
			},
		},
	}

	messageDefs := []Message{
		{
			ID:         "Simple",
			StructName: "Simple",
			Fields:     []Field{},
			Templates: map[string]string{
				"en": "Hello World",
			},
		},
	}

	err := RenderGoI18n(
		outputFile,
		"minimalpkg",
		"en",
		messages,
		[]PlaceholderTemplate{},
		[]Placeholder{},
		messageDefs,
		[]string{"en"},
	)

	s.Assert().NoError(err)
	s.Assert().FileExists(outputFile)

	content, err := os.ReadFile(outputFile)
	s.Require().NoError(err)

	contentStr := string(content)
	s.Assert().Contains(contentStr, "package minimalpkg")
	s.Assert().Contains(contentStr, "Simple")
}

// Unit tests for template functions

func TestTemplateFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name:     "camelCase function",
			template: `{{.name | camelCase}}`,
			data:     map[string]string{"name": "user_name"},
			expected: "userName",
		},
		{
			name:     "title function",
			template: `{{.name | title}}`,
			data:     map[string]string{"name": "hello"},
			expected: "Hello",
		},
		{
			name:     "capitalize function",
			template: `{{.name | capitalize}}`,
			data:     map[string]string{"name": "world"},
			expected: "World",
		},
		{
			name:     "commentSafe single line",
			template: `{{.comment | commentSafe}}`,
			data:     map[string]string{"comment": "Simple comment"},
			expected: "Simple comment",
		},
		{
			name:     "commentSafe multi line",
			template: `{{.comment | commentSafe}}`,
			data:     map[string]string{"comment": "Line 1\nLine 2\nLine 3"},
			expected: "Line 1\n//         Line 2\n//         Line 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executeTemplateDirectly(tt.template, tt.data)
			require.NoError(t, err)

			resultStr := strings.TrimSpace(result)
			assert.Equal(t, tt.expected, resultStr)
		})
	}
}

func TestTemplateFunctions_SortingFunctions(t *testing.T) {
	t.Run("sortLocales", func(t *testing.T) {
		template := `{{range sortLocales .templates}}{{.}},{{end}}`
		data := map[string]interface{}{
			"templates": map[string]string{
				"ja": "こんにちは",
				"en": "Hello",
				"fr": "Bonjour",
			},
		}

		result, err := executeTemplateDirectly(template, data)
		require.NoError(t, err)

		// Should be sorted alphabetically
		assert.Equal(t, "en,fr,ja,", strings.TrimSpace(result))
	})

	t.Run("sortMapKeys", func(t *testing.T) {
		template := `{{range sortMapKeys .data}}{{.}},{{end}}`
		data := map[string]interface{}{
			"data": map[string]map[string]string{
				"zebra":  {"en": "Zebra"},
				"apple":  {"en": "Apple"},
				"banana": {"en": "Banana"},
			},
		}

		result, err := executeTemplateDirectly(template, data)
		require.NoError(t, err)

		// Should be sorted alphabetically
		assert.Equal(t, "apple,banana,zebra,", strings.TrimSpace(result))
	})

	t.Run("lastKey", func(t *testing.T) {
		template := `{{lastKey .templates}}`
		data := map[string]interface{}{
			"templates": map[string]string{
				"ja": "こんにちは",
				"en": "Hello",
				"fr": "Bonjour",
			},
		}

		result, err := executeTemplateDirectly(template, data)
		require.NoError(t, err)

		// Should return the last key alphabetically
		assert.Equal(t, "ja", strings.TrimSpace(result))
	})
}

// createTestFuncMap returns the actual production function map for testing
func createTestFuncMap() template.FuncMap {
	return CreateFuncMap()
}

// Helper function to execute template without Go code formatting
func executeTemplateDirectly(tmplContent string, data interface{}) (string, error) {
	// Create a simplified function map for testing
	funcMap := createTestFuncMap()

	tmpl, err := template.New("test").Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func TestRenderWithConfig_InvalidTemplate(t *testing.T) {
	// Test with invalid template syntax
	invalidTemplate := "{{.invalid syntax"

	_, err := RenderTemplateWithConfig(invalidTemplate, map[string]string{}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse Go template")
}

func TestRenderWithConfig_InvalidGoCode(t *testing.T) {
	// Test with template that generates invalid Go code
	invalidGoTemplate := "package invalid\nfunc { invalid syntax }"

	_, err := RenderTemplateWithConfig(invalidGoTemplate, map[string]string{}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to format generated Go code")
}

func (s *TemplatexTestSuite) TestCreateFuncMapFunctions() {
	funcMap := CreateFuncMap()

	// Test all expected functions are present
	expectedFuncs := []string{
		"sortMapKeys", "sortLocales", "camelCase", "safeIdent",
		"formatPluralTemplate", "title", "capitalize", "commentSafe", "lastKey",
	}

	for _, funcName := range expectedFuncs {
		s.Contains(funcMap, funcName, "Function %s should be in funcMap", funcName)
	}
}

func (s *TemplatexTestSuite) TestSafeIdentFunction() {
	funcMap := CreateFuncMap()
	safeIdentFunc := funcMap["safeIdent"]

	tests := []struct {
		input    string
		expected string
	}{
		{"ValidName", "ValidName"},
		{"123Invalid", "123Invalid"}, // Actual behavior from utils.SafeGoIdentifier
		{"type", "type_"},            // Actual behavior from utils.SafeGoIdentifier
		{"func", "func_"},            // Actual behavior from utils.SafeGoIdentifier
		{"valid_name", "valid_name"},
		{"", ""}, // Actual behavior from utils.SafeGoIdentifier
	}

	for _, tt := range tests {
		result := safeIdentFunc.(func(string) string)(tt.input)
		s.Equal(tt.expected, result, "safeIdent(%s)", tt.input)
	}
}

func (s *TemplatexTestSuite) TestTemplateFunctionEdgeCases() {
	// Test edge cases for template functions to improve coverage
	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name:     "titleFunc with empty string",
			template: `{{.value | title}}`,
			data:     map[string]string{"value": ""},
			expected: "",
		},
		{
			name:     "capitalizeFunc with empty string",
			template: `{{.value | capitalize}}`,
			data:     map[string]string{"value": ""},
			expected: "",
		},
		{
			name:     "camelCase with empty parts",
			template: `{{.value | camelCase}}`,
			data:     map[string]string{"value": "hello__world"},
			expected: "helloWorld",
		},
		{
			name:     "lastKey with empty map",
			template: `{{lastKey .templates}}`,
			data:     map[string]interface{}{"templates": map[string]string{}},
			expected: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := executeTemplateDirectly(tt.template, tt.data)
			s.NoError(err)
			s.Equal(tt.expected, strings.TrimSpace(result))
		})
	}
}

func (s *TemplatexTestSuite) TestFormatPluralTemplateFunction() {
	// Test formatPluralTemplate function coverage
	funcMap := CreateFuncMap()
	formatPluralFunc := funcMap["formatPluralTemplate"]

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string input",
			input:    "Hello {{.name}}",
			expected: `"Hello {{.name}}"`,
		},
		{
			name: "map[string]interface{} with single form",
			input: map[string]interface{}{
				"other": "{{.Count}} items",
			},
			expected: `{other: "{{.Count}} items"}`,
		},
		{
			name: "map[string]interface{} with multiple forms",
			input: map[string]interface{}{
				"one":   "{{.Count}} item",
				"other": "{{.Count}} items",
			},
			expected: `{` + "\n" + `//       one: "{{.Count}} item",` + "\n" + `//       other: "{{.Count}} items"` + "\n" + `//     }`,
		},
		{
			name: "map[interface{}]interface{} input",
			input: map[interface{}]interface{}{
				"one":   "{{.Count}} item",
				"other": "{{.Count}} items",
			},
			expected: `{` + "\n" + `//       one: "{{.Count}} item",` + "\n" + `//       other: "{{.Count}} items"` + "\n" + `//     }`,
		},
		{
			name:     "non-string, non-map input",
			input:    123,
			expected: `"123"`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := formatPluralFunc.(func(interface{}) string)(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *TemplatexTestSuite) TestRenderTemplateWithConfigErrors() {
	// Test error cases for RenderTemplateWithConfig
	tests := []struct {
		name         string
		tmplContent  string
		data         interface{}
		expectError  bool
		errorContains string
	}{
		{
			name:         "template parse error",
			tmplContent:  "{{.invalid syntax",
			data:         map[string]string{},
			expectError:  true,
			errorContains: "failed to parse Go template",
		},
		{
			name:         "template execution error",
			tmplContent:  "package test\n{{call .nonexistent}}",
			data:         map[string]string{},
			expectError:  true,
			errorContains: "failed to execute Go template",
		},
		{
			name:         "invalid Go code generation",
			tmplContent:  "package test\nfunc { invalid syntax }",
			data:         map[string]string{},
			expectError:  true,
			errorContains: "failed to format generated Go code",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := RenderTemplateWithConfig(tt.tmplContent, tt.data, nil)
			if tt.expectError {
				s.Error(err)
				s.Contains(err.Error(), tt.errorContains)
			} else {
				s.NoError(err)
			}
		})
	}
}
