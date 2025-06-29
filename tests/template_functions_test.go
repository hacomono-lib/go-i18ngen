package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/generator"
)

func TestTemplateFunctions(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test template functions for go-i18n backend
	configContent := `
backend: go-i18n
compound: true
locales: [ja, en]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "i18n"
`
	
	// Create directories
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		t.Fatalf("Failed to create messages dir: %v", err)
	}
	if err := os.MkdirAll(placeholdersDir, 0755); err != nil {
		t.Fatalf("Failed to create placeholders dir: %v", err)
	}
	
	// Create config file
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	
	// Create message file with various template functions
	messageContent := `
FormattingTest:
  ja: "{{.status}}の{{.type | upper}}メッセージです"
  en: "This is a {{.type | title}} message with {{.status | lower}} status"

ChainedFunctions:
  ja: "{{.text | title}}の処理"
  en: "Processing {{.text | title | lower}}"

MultipleFields:
  ja: "{{.field:input}}から{{.field:output | upper}}へ"
  en: "From {{.field:input | title}} to {{.field:output | upper}}"
`
	
	messagePath := filepath.Join(messagesDir, "messages.yaml")
	if err := os.WriteFile(messagePath, []byte(messageContent), 0644); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}
	
	// Create placeholder files
	statusContent := `
pending:
  ja: 保留中
  en: pending
completed:
  ja: 完了
  en: completed
`
	
	statusPath := filepath.Join(placeholdersDir, "status.yaml")
	if err := os.WriteFile(statusPath, []byte(statusContent), 0644); err != nil {
		t.Fatalf("Failed to write status placeholders: %v", err)
	}
	
	typeContent := `
info:
  ja: 情報
  en: info
error:
  ja: エラー
  en: error
`
	
	typePath := filepath.Join(placeholdersDir, "type.yaml")
	if err := os.WriteFile(typePath, []byte(typeContent), 0644); err != nil {
		t.Fatalf("Failed to write type placeholders: %v", err)
	}
	
	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	
	// Load config and generate
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if err := generator.Run(cfg); err != nil {
		t.Fatalf("Generation failed: %v", err)
	}
	
	// Verify generated file exists
	outputPath := filepath.Join(tempDir, "i18n.gen.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Generated file not found: %s", outputPath)
	}
	
	// Read and verify generated content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	
	generatedContent := string(content)
	
	t.Run("template function metadata", func(t *testing.T) {
		// Should have templateFunctions variable
		if !strings.Contains(generatedContent, "var templateFunctions = map[string]map[string]map[string][]string{") {
			t.Error("Should generate templateFunctions metadata")
		}
		
		// Should contain function metadata for specific cases
		if !strings.Contains(generatedContent, `"FormattingTest":`) {
			t.Error("Should contain FormattingTest in templateFunctions")
		}
	})
	
	t.Run("function application", func(t *testing.T) {
		// Should have processField function
		if !strings.Contains(generatedContent, "func processField") {
			t.Error("Should generate processField function")
		}
		
		// Should have applyTemplateFunctions function
		if !strings.Contains(generatedContent, "func applyTemplateFunctions") {
			t.Error("Should generate applyTemplateFunctions function")
		}
		
		// Should support title, upper, lower functions
		if !strings.Contains(generatedContent, `case "title":`) ||
		   !strings.Contains(generatedContent, `case "upper":`) ||
		   !strings.Contains(generatedContent, `case "lower":`) {
			t.Error("Should support title, upper, lower functions in applyTemplateFunctions")
		}
	})
	
	t.Run("field processing", func(t *testing.T) {
		// Should call processField in buildTemplateData
		if !strings.Contains(generatedContent, "processField(value, messageID, fieldName, locale)") {
			t.Error("Should call processField in buildTemplateData")
		}
	})
	
	t.Run("generated structures", func(t *testing.T) {
		// Should generate structs with proper field names
		if !strings.Contains(generatedContent, "type FormattingTest struct") {
			t.Error("Should generate FormattingTest struct")
		}
		
		if !strings.Contains(generatedContent, "type MultipleFields struct") {
			t.Error("Should generate MultipleFields struct") 
		}
		
		// Should generate proper field names for suffix notation
		if !strings.Contains(generatedContent, "FieldInput") ||
		   !strings.Contains(generatedContent, "FieldOutput") {
			t.Error("Should generate proper field names for suffix notation")
		}
	})
	
	t.Logf("Template functions test passed")
}

func TestBuiltinTemplateFunctions(t *testing.T) {
	// Test that builtin backend also handles template functions correctly
	tempDir := t.TempDir()
	
	configContent := `
backend: builtin
compound: true
locales: [ja, en]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "i18n"
`
	
	// Create directories
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		t.Fatalf("Failed to create messages dir: %v", err)
	}
	if err := os.MkdirAll(placeholdersDir, 0755); err != nil {
		t.Fatalf("Failed to create placeholders dir: %v", err)
	}
	
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	
	// Create message file with template functions
	messageContent := `
TitleTest:
  ja: "{{.name | title}}さんの情報"
  en: "{{.name | title}}'s Information"
`
	
	messagePath := filepath.Join(messagesDir, "messages.yaml")
	if err := os.WriteFile(messagePath, []byte(messageContent), 0644); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}
	
	nameContent := `
user:
  ja: ユーザー
  en: user
`
	
	namePath := filepath.Join(placeholdersDir, "name.yaml")
	if err := os.WriteFile(namePath, []byte(nameContent), 0644); err != nil {
		t.Fatalf("Failed to write placeholders: %v", err)
	}
	
	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	
	// Load config and generate
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if err := generator.Run(cfg); err != nil {
		t.Fatalf("Generation failed: %v", err)
	}
	
	// Verify generated file exists
	outputPath := filepath.Join(tempDir, "i18n.gen.go")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	
	generatedContent := string(content)
	
	// Builtin should preserve template functions in the template string
	if !strings.Contains(generatedContent, "{{.name | title}}") {
		t.Error("Builtin backend should preserve template functions in templates")
	}
	
	// Should have getTemplateFunctions with title function
	if !strings.Contains(generatedContent, `"title": func(s string) string`) {
		t.Error("Builtin backend should have title function in getTemplateFunctions")
	}
}