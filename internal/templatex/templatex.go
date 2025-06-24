// Package templatex handles Go template rendering and code generation for i18n files.
package templatex

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"
	"text/template"
)

//go:embed i18n.gotmpl
var templateContent string

type Message struct {
	ID         string
	StructName string
	Fields     []Field
	Templates  map[string]string // locale -> template
}

type Field struct {
	FieldName   string
	Type        string
	TemplateKey string
}

type Placeholder struct {
	StructName string
	VarName    string
	IsValue    bool
	Items      []PlaceholderItem
}

type PlaceholderItem struct {
	ID        string
	FieldName string
	Templates map[string]string // locale -> localized value
}

type MessageTemplate struct {
	ID        string
	Templates map[string]string
}

type PlaceholderTemplate struct {
	Name            string
	HasLocaleFiles  bool
	LocaleTemplates map[string]map[string]string
}

type TemplateDef struct {
	PackageName     string
	PrimaryLocale   string
	Messages        []MessageTemplate
	Placeholders    []PlaceholderTemplate
	PlaceholderDefs []Placeholder
	MessageDefs     []Message
}

// TemplateConfig represents configuration for template generation
type TemplateConfig struct {
	// Future configuration options can be added here
}

// CreateFuncMap creates the template function map used for rendering
func CreateFuncMap() template.FuncMap {
	return template.FuncMap{
		"camelCase": func(s string) string {
			parts := strings.Split(s, "_")
			if len(parts) == 0 {
				return s
			}
			// First part stays lowercase, subsequent parts are capitalized
			result := parts[0]
			for i := 1; i < len(parts); i++ {
				if parts[i] != "" {
					result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
				}
			}
			return result
		},
		"title": func(s string) string {
			if s == "" {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"capitalize": func(s string) string {
			if s == "" {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"commentSafe": func(s string) string {
			// Properly format multi-line strings as comments
			lines := strings.Split(s, "\n")
			if len(lines) <= 1 {
				return s
			}

			// For multi-line strings, properly convert newlines to comment format
			var result []string
			for i, line := range lines {
				trimmed := strings.TrimRight(line, "\r")
				if i == 0 {
					result = append(result, trimmed)
				} else {
					// Add proper indentation and comment prefix for lines after the first
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
}

// RenderTemplateWithConfig renders a template with the given data and config
func RenderTemplateWithConfig(tmplContent string, data any, config *TemplateConfig) ([]byte, error) {
	// Use the shared function map
	funcMap := CreateFuncMap()

	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go template: %w", err)
	}

	var buf bytes.Buffer
	if execErr := tmpl.Execute(&buf, data); execErr != nil {
		return nil, fmt.Errorf("failed to execute Go template: %w", execErr)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format generated Go code: %w", err)
	}

	return formatted, nil
}

func Render(
	outPath, pkg, primaryLocale string,
	messages []MessageTemplate,
	placeholders []PlaceholderTemplate,
	placeholderDefs []Placeholder,
	messageDefs []Message,
) error {
	return RenderWithConfig(outPath, pkg, primaryLocale, messages, placeholders, placeholderDefs, messageDefs, nil)
}

func RenderWithConfig(
	outPath, pkg, primaryLocale string,
	messages []MessageTemplate,
	placeholders []PlaceholderTemplate,
	placeholderDefs []Placeholder,
	messageDefs []Message,
	config *TemplateConfig,
) error {
	code, err := RenderTemplateWithConfig(templateContent, TemplateDef{
		PackageName:     pkg,
		PrimaryLocale:   primaryLocale,
		Messages:        messages,
		Placeholders:    placeholders,
		PlaceholderDefs: placeholderDefs,
		MessageDefs:     messageDefs,
	}, config)
	if err != nil {
		return err // Already wrapped with detailed context
	}

	if err := os.WriteFile(outPath, code, 0600); err != nil {
		return fmt.Errorf("failed to write generated code to file %q: %w", outPath, err)
	}

	return nil
}
