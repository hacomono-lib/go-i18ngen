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

	"github.com/hacomono-lib/go-i18ngen/internal/utils"
)

//go:embed go-i18n.gotmpl
var goI18nTemplateContent string

type Message struct {
	ID                string
	StructName        string
	Fields            []Field
	Templates         map[string]string      // locale -> template (simplified for processing)
	RawTemplates      map[string]interface{} // locale -> raw template data (preserves plural forms)
	SupportsCount     bool
	PluralPlaceholder string // The actual plural placeholder key used (e.g., "Count", "Quantity")
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
	PackageName      string
	PrimaryLocale    string
	Messages         []MessageTemplate
	Placeholders     []PlaceholderTemplate
	PlaceholderDefs  []Placeholder
	MessageDefs      []Message
	Locales          []string
	MessagesByLocale map[string]map[string]string
}

// TemplateConfig represents configuration for template generation
type TemplateConfig struct {
	// Future configuration options can be added here
}

// Helper functions

// convertRawTemplateToYaml converts a raw template (which may be string or map) to YAML format
func convertRawTemplateToYaml(rawTemplate interface{}) string {
	switch v := rawTemplate.(type) {
	case string:
		// Simple string template - wrap in quotes and add space
		return " \"" + strings.ReplaceAll(v, "\"", "\\\"") + "\""
	case map[string]interface{}:
		// Plural forms map (e.g., {"one": "...", "other": "..."})
		// Convert to YAML block format for go-i18n
		var parts []string
		for form, template := range v {
			if tmpl, ok := template.(string); ok {
				parts = append(parts, fmt.Sprintf("%s: %q", form, tmpl))
			}
		}
		sort.Strings(parts) // Consistent ordering
		if len(parts) > 0 {
			// For plural forms in go-i18n YAML format
			return "\n  " + strings.Join(parts, "\n  ")
		}
		return ""
	case map[interface{}]interface{}:
		// Handle YAML-parsed format
		var parts []string
		for k, v := range v {
			if form, ok := k.(string); ok {
				if tmpl, ok := v.(string); ok {
					parts = append(parts, fmt.Sprintf("%s: %q", form, tmpl))
				}
			}
		}
		sort.Strings(parts)
		if len(parts) > 0 {
			// For plural forms in go-i18n YAML format
			return "\n  " + strings.Join(parts, "\n  ")
		}
		return ""
	default:
		// Unknown type, convert to string
		return fmt.Sprintf("%v", v)
	}
}

// findMessageDef finds a MessageDef by ID
func findMessageDef(messageDefs []Message, id string) *Message {
	for i, msgDef := range messageDefs {
		if msgDef.ID == id {
			return &messageDefs[i]
		}
	}
	return nil
}

// Template function implementations

func camelCaseFunc(s string) string {
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
}

func titleFunc(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func capitalizeFunc(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func commentSafeFunc(s string) string {
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
}

func sortLocalesFunc(templates map[string]string) []string {
	var locales []string
	for locale := range templates {
		locales = append(locales, locale)
	}
	sort.Strings(locales)
	return locales
}

func sortMapKeysFunc(m map[string]map[string]string) []string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func lastKeyFunc(m map[string]string) string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return ""
	}
	sort.Strings(keys)
	return keys[len(keys)-1]
}

func formatPluralTemplateFunc(tmpl interface{}) string {
	switch t := tmpl.(type) {
	case string:
		return fmt.Sprintf("%q", t)
	case map[string]interface{}:
		return formatStringInterfaceMap(t)
	case map[interface{}]interface{}:
		return formatInterfaceMap(t)
	default:
		return fmt.Sprintf("\"%v\"", t)
	}
}

func formatStringInterfaceMap(t map[string]interface{}) string {
	if len(t) <= 1 {
		// Single form, keep it inline
		for form, text := range t {
			return fmt.Sprintf("{%s: %q}", form, text)
		}
	}
	// Multiple forms, format with newlines for readability
	var parts []string
	for form, text := range t {
		parts = append(parts, fmt.Sprintf("//       %s: %q", form, text))
	}
	sort.Strings(parts) // Sort for consistent output
	if len(parts) > 1 {
		return "{\n" + strings.Join(parts, ",\n") + "\n//     }"
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func formatInterfaceMap(t map[interface{}]interface{}) string {
	// Handle YAML-parsed maps with interface{} keys
	var parts []string
	for key, text := range t {
		if keyStr, ok := key.(string); ok {
			parts = append(parts, fmt.Sprintf("//       %s: %q", keyStr, text))
		}
	}
	sort.Strings(parts)
	if len(parts) > 1 {
		return "{\n" + strings.Join(parts, ",\n") + "\n//     }"
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// CreateFuncMap creates the template function map used for rendering
func CreateFuncMap() template.FuncMap {
	return template.FuncMap{
		"camelCase":            camelCaseFunc,
		"title":                titleFunc,
		"capitalize":           capitalizeFunc,
		"commentSafe":          commentSafeFunc,
		"sortLocales":          sortLocalesFunc,
		"sortMapKeys":          sortMapKeysFunc,
		"lastKey":              lastKeyFunc,
		"formatPluralTemplate": formatPluralTemplateFunc,
		"safeIdent":            utils.SafeGoIdentifier,
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

func RenderGoI18n(
	outPath, pkg, primaryLocale string,
	messages []MessageTemplate,
	placeholders []PlaceholderTemplate,
	placeholderDefs []Placeholder,
	messageDefs []Message,
	locales []string,
) error {
	return RenderGoI18nWithConfig(outPath, pkg, primaryLocale, messages, placeholders, placeholderDefs, messageDefs, locales, nil)
}

func RenderGoI18nWithConfig(
	outPath, pkg, primaryLocale string,
	messages []MessageTemplate,
	placeholders []PlaceholderTemplate,
	placeholderDefs []Placeholder,
	messageDefs []Message,
	locales []string,
	config *TemplateConfig,
) error {
	// Build message data by locale for go-i18n
	messagesByLocale := make(map[string]map[string]string)
	for _, locale := range locales {
		messagesByLocale[locale] = make(map[string]string)
	}

	// Use both RawTemplates (for plural forms) and processed Templates (for suffix notation)
	for _, msgDef := range messageDefs {
		if msgDef.RawTemplates != nil {
			// Check if this is a plural message by looking for map structures
			for locale, rawTemplate := range msgDef.RawTemplates {
				if messagesByLocale[locale] == nil {
					messagesByLocale[locale] = make(map[string]string)
				}

				// If rawTemplate is a map, it's a plural form - use it directly
				switch rawTemplate.(type) {
				case map[string]interface{}, map[interface{}]interface{}:
					messagesByLocale[locale][msgDef.ID] = convertRawTemplateToYaml(rawTemplate)
				default:
					// For non-plural templates, use processed Templates to get suffix notation conversion
					if processedTemplate, exists := msgDef.Templates[locale]; exists {
						messagesByLocale[locale][msgDef.ID] = convertRawTemplateToYaml(processedTemplate)
					} else {
						// Fallback to raw template if processed version not available
						messagesByLocale[locale][msgDef.ID] = convertRawTemplateToYaml(rawTemplate)
					}
				}
			}
		} else {
			// Use processed Templates if RawTemplates not available
			for locale, template := range msgDef.Templates {
				if messagesByLocale[locale] == nil {
					messagesByLocale[locale] = make(map[string]string)
				}
				messagesByLocale[locale][msgDef.ID] = convertRawTemplateToYaml(template)
			}
		}
	}

	// Also add any messages that don't have MessageDef equivalent
	for _, msg := range messages {
		if msgDef := findMessageDef(messageDefs, msg.ID); msgDef == nil {
			for locale, template := range msg.Templates {
				if messagesByLocale[locale] == nil {
					messagesByLocale[locale] = make(map[string]string)
				}
				messagesByLocale[locale][msg.ID] = template
			}
		}
	}

	code, err := RenderTemplateWithConfig(goI18nTemplateContent, TemplateDef{
		PackageName:      pkg,
		PrimaryLocale:    primaryLocale,
		Messages:         messages,
		Placeholders:     placeholders,
		PlaceholderDefs:  placeholderDefs,
		MessageDefs:      messageDefs,
		Locales:          locales,
		MessagesByLocale: messagesByLocale,
	}, config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outPath, code, 0600); err != nil {
		return fmt.Errorf("failed to write generated code to file %q: %w", outPath, err)
	}

	return nil
}
