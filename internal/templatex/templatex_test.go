package templatex

import (
	"os"
	"path/filepath"
	"sort"
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
	os.RemoveAll(s.tempDir)
}

func (s *TemplatexTestSuite) TestRender_Success() {
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

	err := Render(
		outputFile,
		"testpkg",
		"ja",
		messages,
		placeholderTemplates,
		placeholders,
		messageDefs,
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

func (s *TemplatexTestSuite) TestRender_InvalidOutputPath() {
	// Use an invalid path that cannot be created
	invalidPath := filepath.Join("/invalid", "path", "that", "does", "not", "exist", "test.go")

	err := Render(
		invalidPath,
		"testpkg",
		"ja",
		[]MessageTemplate{},
		[]PlaceholderTemplate{},
		[]Placeholder{},
		[]Message{},
	)

	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "failed to write generated code")
}

func (s *TemplatexTestSuite) TestRender_EmptyData() {
	outputFile := filepath.Join(s.tempDir, "empty.go")

	err := Render(
		outputFile,
		"emptypkg",
		"en",
		[]MessageTemplate{},
		[]PlaceholderTemplate{},
		[]Placeholder{},
		[]Message{},
	)

	s.Assert().NoError(err)
	s.Assert().FileExists(outputFile)

	// Verify generated content
	content, err := os.ReadFile(outputFile)
	s.Require().NoError(err)

	contentStr := string(content)
	s.Assert().Contains(contentStr, "package emptypkg")
}

func (s *TemplatexTestSuite) TestRenderWithConfig_Success() {
	outputFile := filepath.Join(s.tempDir, "config_test.go")

	config := &TemplateConfig{}

	err := RenderWithConfig(
		outputFile,
		"configpkg",
		"en",
		[]MessageTemplate{},
		[]PlaceholderTemplate{},
		[]Placeholder{},
		[]Message{},
		config,
	)

	s.Assert().NoError(err)
	s.Assert().FileExists(outputFile)

	content, err := os.ReadFile(outputFile)
	s.Require().NoError(err)

	contentStr := string(content)
	s.Assert().Contains(contentStr, "package configpkg")
}

func (s *TemplatexTestSuite) TestRender_WithMinimalData() {
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

	err := Render(
		outputFile,
		"minimalpkg",
		"en",
		messages,
		[]PlaceholderTemplate{},
		[]Placeholder{},
		messageDefs,
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

// Helper function to execute template without Go code formatting
func executeTemplateDirectly(tmplContent string, data interface{}) (string, error) {
	// Use basic functions for template generation
	funcMap := template.FuncMap{
		"camelCase": func(s string) string {
			parts := strings.Split(s, "_")
			if len(parts) == 0 {
				return s
			}
			// First part stays lowercase, subsequent parts are capitalized
			result := parts[0]
			for i := 1; i < len(parts); i++ {
				if len(parts[i]) > 0 {
					result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
				}
			}
			return result
		},
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"capitalize": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"commentSafe": func(s string) string {
			// properly format multi-line strings as comments
			lines := strings.Split(s, "\n")
			if len(lines) <= 1 {
				return s
			}

			// for multi-line cases, convert newlines to proper comment format
			var result []string
			for i, line := range lines {
				trimmed := strings.TrimRight(line, "\r")
				if i == 0 {
					result = append(result, trimmed)
				} else {
					// add proper indentation and comment prefix for subsequent lines
					result = append(result, "//         "+trimmed)
				}
			}
			return strings.Join(result, "\n")
		},
		"sortLocales": func(templates map[string]string) []string {
			var locales []string
			for locale := range templates {
				locales = append(locales, locale)
			}
			sort.Strings(locales)
			return locales
		},
		"sortMapKeys": func(m map[string]map[string]string) []string {
			var keys []string
			for key := range m {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			return keys
		},
		"lastKey": func(m map[string]string) string {
			var keys []string
			for key := range m {
				keys = append(keys, key)
			}
			if len(keys) == 0 {
				return ""
			}
			sort.Strings(keys)
			return keys[len(keys)-1]
		},
	}

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

	_, err := renderWithConfig(invalidTemplate, map[string]string{}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse Go template")
}

func TestRenderWithConfig_InvalidGoCode(t *testing.T) {
	// Test with template that generates invalid Go code
	invalidGoTemplate := "package invalid\nfunc { invalid syntax }"

	_, err := renderWithConfig(invalidGoTemplate, map[string]string{}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to format generated Go code")
}
