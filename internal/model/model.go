package model

import (
	"regexp"
	"sort"

	"github.com/hacomono-lib/go-i18ngen/internal/templatex"
	"github.com/hacomono-lib/go-i18ngen/internal/utils"
)

// Pre-compiled regular expressions for better performance
var (
	digitStartPattern = regexp.MustCompile(`^[0-9]`)
)

// FieldInfo represents a field with optional suffix for enhanced naming
type FieldInfo struct {
	Name   string // Base field name (e.g., "entity")
	Suffix string // Optional suffix (e.g., "from", "1", "user")
}

// String returns the field identifier for template processing
func (f FieldInfo) String() string {
	if f.Suffix != "" {
		return f.Name + ":" + f.Suffix
	}
	return f.Name
}

// GenerateFieldName creates the Go struct field name
func (f FieldInfo) GenerateFieldName() string {
	if f.Suffix != "" {
		return utils.ToCamelCase(f.Name) + utils.ToCamelCase(f.Suffix)
	}
	return utils.ToCamelCase(f.Name)
}

// GenerateTemplateKey creates the template key for rendering
func (f FieldInfo) GenerateTemplateKey() string {
	if f.Suffix != "" {
		return f.Name + utils.ToCamelCase(f.Suffix)
	}
	return f.Name
}

type MessageSource struct {
	ID         string
	Templates  map[string]string // locale -> template
	FieldInfos []FieldInfo       // Enhanced field information with suffix support
}

type PlaceholderSource struct {
	Kind  string
	Items map[string]map[string]string // ID -> locale -> string
}

type Definitions struct {
	Messages     []templatex.Message
	Placeholders []templatex.Placeholder
}

// generateStructName generates a valid Go struct name from a message ID
// If the ID starts with a digit, it prefixes with "Msg"
func generateStructName(id string) string {
	// Check if ID starts with a digit
	if digitStartPattern.MatchString(id) {
		return "Msg" + utils.ToCamelCase(id)
	}
	return utils.ToCamelCase(id)
}

func Build(messages []MessageSource, placeholders []PlaceholderSource, locales []string) (*Definitions, error) {
	defs := Definitions{}

	// Determine primary locale (first locale in configuration)
	primaryLocale := "en" // Default fallback
	if len(locales) > 0 {
		primaryLocale = locales[0]
	}

	// Build placeholder definitions
	placeholderTypes := map[string]string{}
	for _, ph := range placeholders {
		// Determine if it's a Value placeholder (no localization)
		isValue := true
		for _, localeMap := range ph.Items {
			if len(localeMap) > 0 {
				isValue = false
				break
			}
		}

		// Generate type name based on whether it has localization
		var typeName string
		if isValue {
			typeName = utils.ToCamelCase(ph.Kind) + "Value"
		} else {
			typeName = utils.ToCamelCase(ph.Kind) + "Text"
		}
		varName := ph.Kind + "Templates"

		// Generate items for utility access
		var items []templatex.PlaceholderItem
		for id, templates := range ph.Items {
			items = append(items, templatex.PlaceholderItem{
				ID:        id,
				FieldName: utils.ToCamelCase(id),
				Templates: templates,
			})
		}

		// Sort items by their localized text in primary locale for consistent ordering
		if !isValue {
			sort.Slice(items, func(i, j int) bool {
				// Get localized text for primary locale, fallback to ID if not available
				textI := ph.Items[items[i].ID][primaryLocale]
				if textI == "" {
					textI = items[i].ID
				}
				textJ := ph.Items[items[j].ID][primaryLocale]
				if textJ == "" {
					textJ = items[j].ID
				}
				return textI < textJ
			})
		}

		defs.Placeholders = append(defs.Placeholders, templatex.Placeholder{
			StructName: typeName,
			VarName:    varName,
			IsValue:    isValue,
			Items:      items,
		})

		// Map the kind itself to the type (for {{.entity}} usage)
		placeholderTypes[ph.Kind] = typeName

		// Also map individual items (for {{.user}} usage)
		for id := range ph.Items {
			placeholderTypes[id] = typeName
		}
	}

	// Build message definitions
	for _, msg := range messages {
		structName := generateStructName(msg.ID)
		var fields []templatex.Field

		// Process FieldInfos to generate fields
		for _, fieldInfo := range msg.FieldInfos {
			fieldName := fieldInfo.GenerateFieldName()
			templateKey := fieldInfo.GenerateTemplateKey()

			// Determine the base field name for type lookup
			baseFieldName := fieldInfo.Name
			typ, ok := placeholderTypes[baseFieldName]
			if !ok {
				// Field not found in placeholder definitions, treat as Value type
				typ = utils.ToCamelCase(baseFieldName) + "Value"

				// Add to placeholder definitions if not already present
				placeholderAlreadyExists := false
				for _, ph := range defs.Placeholders {
					if ph.StructName == typ {
						placeholderAlreadyExists = true
						break
					}
				}
				if !placeholderAlreadyExists {
					// For auto-generated value types, create single item
					items := []templatex.PlaceholderItem{{
						ID:        baseFieldName,
						FieldName: utils.ToCamelCase(baseFieldName),
						Templates: make(map[string]string), // Empty templates for Value types
					}}

					defs.Placeholders = append(defs.Placeholders, templatex.Placeholder{
						StructName: typ,
						VarName:    baseFieldName + "Templates",
						IsValue:    true,
						Items:      items,
					})
				}
			}

			fields = append(fields, templatex.Field{
				FieldName:   fieldName,
				Type:        typ,
				TemplateKey: templateKey,
			})
		}

		// Process templates to handle suffix-based or duplicate placeholders
		originalTemplates := msg.Templates
		if originalTemplates == nil {
			originalTemplates = make(map[string]string)
		}

		// Process templates with FieldInfos
		processedTemplates := ProcessMessageTemplatesWithFieldInfos(originalTemplates, msg.FieldInfos)

		defs.Messages = append(defs.Messages, templatex.Message{
			ID:         msg.ID,
			StructName: structName,
			Fields:     fields,
			Templates:  processedTemplates,
		})
	}

	// Sort for consistent output (CI-friendly)
	sort.Slice(defs.Messages, func(i, j int) bool {
		return defs.Messages[i].ID < defs.Messages[j].ID
	})

	sort.Slice(defs.Placeholders, func(i, j int) bool {
		return defs.Placeholders[i].StructName < defs.Placeholders[j].StructName
	})

	return &defs, nil
}

func BuildTemplates(messages []MessageSource, placeholders []PlaceholderSource, locales []string) ([]templatex.MessageTemplate, []templatex.PlaceholderTemplate, error) {
	var messageTemplates []templatex.MessageTemplate
	var placeholderTemplates []templatex.PlaceholderTemplate

	// Build message templates
	for _, msg := range messages {
		templates := msg.Templates
		if templates == nil {
			templates = make(map[string]string)
		}
		messageTemplates = append(messageTemplates, templatex.MessageTemplate{
			ID:        msg.ID,
			Templates: templates,
		})
	}

	// Build placeholder templates
	for _, ph := range placeholders {
		hasLocaleFiles := len(ph.Items) > 0
		for _, localeMap := range ph.Items {
			if len(localeMap) > 0 {
				hasLocaleFiles = true
				break
			}
		}

		if hasLocaleFiles {
			placeholderTemplates = append(placeholderTemplates, templatex.PlaceholderTemplate{
				Name:            ph.Kind,
				HasLocaleFiles:  true,
				LocaleTemplates: ph.Items,
			})
		}
	}

	// Sort for consistent output (CI-friendly)
	sort.Slice(messageTemplates, func(i, j int) bool {
		return messageTemplates[i].ID < messageTemplates[j].ID
	})

	sort.Slice(placeholderTemplates, func(i, j int) bool {
		return placeholderTemplates[i].Name < placeholderTemplates[j].Name
	})

	return messageTemplates, placeholderTemplates, nil
}
