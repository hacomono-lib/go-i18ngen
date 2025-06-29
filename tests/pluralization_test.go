package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/generator"
)

var (
	// Pre-compiled regex patterns for better performance
	countFieldPattern = regexp.MustCompile(`count\s+\*int`)
)

func TestPluralizationSupport(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test pluralization for go-i18n backend
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
	
	// Create message file with pluralization
	messageContent := `
ItemCount:
  ja: "{{.entity}} アイテム ({{.Count}}個)"
  en:
    one: "{{.entity}} item"
    other: "{{.entity}} items ({{.Count}})"

UserCount:
  ja: "{{.Count}}人のユーザー"
  en:
    one: "{{.Count}} user"
    other: "{{.Count}} users"

FileCount:
  ja: "{{.Number}}個のファイル"
  en:
    one: "{{.Number}} file"
    other: "{{.Number}} files"

RegularMessage:
  ja: "通常のメッセージです"
  en: "This is a regular message"
`
	
	messagePath := filepath.Join(messagesDir, "messages.yaml")
	if err := os.WriteFile(messagePath, []byte(messageContent), 0644); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}
	
	// Create placeholder files
	entityContent := `
product:
  ja: 製品
  en: Product
file:
  ja: ファイル
  en: File
`
	
	entityPath := filepath.Join(placeholdersDir, "entity.yaml")
	if err := os.WriteFile(entityPath, []byte(entityContent), 0644); err != nil {
		t.Fatalf("Failed to write entity placeholders: %v", err)
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
	
	t.Run("pluralization detection", func(t *testing.T) {
		// Messages with {{.Count}} should have WithCount support
		checkWithCountSupport := func(msgName string, shouldHave bool) {
			structDecl := "type " + msgName + " struct"
			structPos := strings.Index(generatedContent, structDecl)
			if structPos == -1 {
				t.Errorf("Struct %s not found", msgName)
				return
			}
			
			// Find the end of the struct for field checking
			nextTypePos := strings.Index(generatedContent[structPos+1:], "\ntype ")
			var structContent string
			if nextTypePos == -1 {
				structContent = generatedContent[structPos:]
			} else {
				structContent = generatedContent[structPos : structPos+nextTypePos]
			}
			
			// Look for "count *int" field declaration using pre-compiled regex
			hasCountField := countFieldPattern.MatchString(structContent)
			// WithCount method is defined outside the struct, so search in full content
			withCountPattern := "func (m " + msgName + ") WithCount(count int)"
			hasWithCount := strings.Contains(generatedContent, withCountPattern)
			
			if shouldHave {
				if !hasCountField {
					t.Errorf("%s should have count field for pluralization", msgName)
				}
				if !hasWithCount {
					t.Errorf("%s should have WithCount method", msgName)
				}
			} else {
				if hasCountField {
					t.Errorf("%s should not have count field", msgName)
				}
				if hasWithCount {
					t.Errorf("%s should not have WithCount method", msgName)
				}
			}
		}
		
		// These should have pluralization support
		checkWithCountSupport("ItemCount", true)
		checkWithCountSupport("UserCount", true)
		checkWithCountSupport("FileCount", true)
		
		// This should not have pluralization support
		checkWithCountSupport("RegularMessage", false)
	})
	
	t.Run("pluralization config", func(t *testing.T) {
		// Should set PluralCount in LocalizeConfig
		if !strings.Contains(generatedContent, "config.PluralCount = *m.count") {
			t.Error("Should set PluralCount in LocalizeConfig")
		}
		
		// Should add Count to TemplateData
		if !strings.Contains(generatedContent, `config.TemplateData["Count"] = *m.count`) {
			t.Error("Should add Count to TemplateData")
		}
	})
	
	t.Run("various count placeholders", func(t *testing.T) {
		// Should NOT generate Value types for plural placeholders (Count, Number)
		// These are handled by WithCount() method instead
		countTypes := []string{"Count", "Number"}
		for _, countType := range countTypes {
			if strings.Contains(generatedContent, countType+"Value") {
				t.Errorf("Should NOT generate %sValue type for plural placeholder", countType)
			}
		}
	})
	
	t.Run("plural template parsing", func(t *testing.T) {
		// English templates should be converted from plural forms
		// ItemCount should use "other" form: "{{.entity}} items ({{.Count}})"
		if !strings.Contains(generatedContent, `ItemCount: "{{.entity}} items ({{.Count}})"`) {
			t.Error("Should convert English plural to 'other' form for ItemCount")
		}
		
		// UserCount should use "other" form: "{{.Count}} users"
		if !strings.Contains(generatedContent, `UserCount: "{{.Count}} users"`) {
			t.Error("Should convert English plural to 'other' form for UserCount")
		}
	})
	
	t.Logf("Pluralization test passed")
}

func TestPluralizationParser(t *testing.T) {
	// Test the parser's ability to handle mixed YAML formats
	tempDir := t.TempDir()
	
	configContent := `
backend: go-i18n
compound: true
locales: [en]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "i18n"
`
	
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
	
	// Test various plural formats that should be parsed correctly
	messageContent := `
# Mixed format: string and plural object
MixedFormat:
  en:
    one: "One item"
    other: "{{.Count}} items"

# String only format
StringOnly:
  en: "Simple message"

# Complex plural with multiple forms
ComplexPlural:
  en:
    zero: "No items"
    one: "One item"
    few: "Few items"
    many: "Many items"
    other: "{{.Count}} items"
`
	
	messagePath := filepath.Join(messagesDir, "test.yaml")
	if err := os.WriteFile(messagePath, []byte(messageContent), 0644); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}
	
	// Minimal placeholder to avoid errors
	placeholderContent := `
dummy:
  en: dummy
`
	
	placeholderPath := filepath.Join(placeholdersDir, "dummy.yaml")
	if err := os.WriteFile(placeholderPath, []byte(placeholderContent), 0644); err != nil {
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
	
	// Verify file was generated successfully
	outputPath := filepath.Join(tempDir, "i18n.gen.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Generated file not found: %s", outputPath)
	}
	
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	
	generatedContent := string(content)
	
	// Should generate all message types
	expectedMessages := []string{"MixedFormat", "StringOnly", "ComplexPlural"}
	for _, msg := range expectedMessages {
		if !strings.Contains(generatedContent, "type "+msg+" struct") {
			t.Errorf("Should generate %s struct", msg)
		}
	}
	
	// MixedFormat should have WithCount (has {{.Count}})
	if !strings.Contains(generatedContent, "func (m MixedFormat) WithCount") {
		t.Error("MixedFormat should have WithCount method")
	}
	
	// StringOnly should not have WithCount
	if strings.Contains(generatedContent, "func (m StringOnly) WithCount") {
		t.Error("StringOnly should not have WithCount method")
	}
	
	t.Logf("Pluralization parser test passed")
}