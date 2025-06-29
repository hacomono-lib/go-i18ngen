package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/generator"
)

// TestFinalIntegration tests all the key features implemented
func TestFinalIntegration(t *testing.T) {
	t.Run("BuiltinBackend", func(t *testing.T) {
		testBuiltinBackend(t)
	})
	
	t.Run("GoI18nBackend", func(t *testing.T) {
		testGoI18nBackend(t)
	})
	
	t.Run("MultiLanguageSupport", func(t *testing.T) {
		testMultiLanguageSupport(t)
	})
	
	t.Log("All major features tested successfully!")
}

func testBuiltinBackend(t *testing.T) {
	tempDir := t.TempDir()
	
	// Setup builtin test
	setupTestFiles(t, tempDir, "builtin", map[string]string{
		"TestMessage": `
  ja: "{{.name | title}}さんのテスト"
  en: "Test for {{.name | title}}"`,
	})
	
	// Generate and verify
	if err := runGeneration(t, tempDir); err != nil {
		t.Fatalf("Builtin generation failed: %v", err)
	}
	
	content := readGeneratedFile(t, tempDir)
	
	// Verify builtin-specific features
	if !strings.Contains(content, "text/template") {
		t.Error("Builtin should use text/template")
	}
	if strings.Contains(content, "go-i18n") {
		t.Error("Builtin should not import go-i18n")
	}
	if !strings.Contains(content, "templateCache") {
		t.Error("Builtin should have templateCache")
	}
	
	t.Log("✓ Builtin backend test passed")
}

func testGoI18nBackend(t *testing.T) {
	tempDir := t.TempDir()
	
	// Setup go-i18n test
	setupTestFiles(t, tempDir, "go-i18n", map[string]string{
		"TestMessage": `
  ja: "{{.name}}さんのテスト"
  en: "Test for {{.name}}"`,
		"CountMessage": `
  ja: "{{.Count}}個のアイテム"
  en:
    one: "{{.Count}} item"
    other: "{{.Count}} items"`,
	})
	
	// Generate and verify
	if err := runGeneration(t, tempDir); err != nil {
		t.Fatalf("Go-i18n generation failed: %v", err)
	}
	
	content := readGeneratedFile(t, tempDir)
	
	// Verify go-i18n-specific features
	if !strings.Contains(content, "go-i18n") {
		t.Error("Go-i18n backend should import go-i18n")
	}
	if !strings.Contains(content, "messageData") {
		t.Error("Go-i18n should have embedded messageData")
	}
	if !strings.Contains(content, "placeholderData") {
		t.Error("Go-i18n should have embedded placeholderData")
	}
	if !strings.Contains(content, "WithCount") {
		t.Error("Go-i18n should support WithCount for pluralization")
	}
	
	t.Log("✓ Go-i18n backend test passed")
}

func testMultiLanguageSupport(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test with 5 languages
	setupTestFiles(t, tempDir, "go-i18n", map[string]string{
		"WelcomeMessage": `
  ja: "ようこそ"
  en: "Welcome"
  fr: "Bienvenue"
  de: "Willkommen"
  es: "Bienvenido"`,
	})
	
	// Update config for multiple languages
	configContent := `
backend: go-i18n
compound: true
locales: [ja, en, fr, de, es]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "i18n"
`
	
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	
	if err := runGeneration(t, tempDir); err != nil {
		t.Fatalf("Multi-language generation failed: %v", err)
	}
	
	content := readGeneratedFile(t, tempDir)
	
	// Verify all languages are included
	languages := []string{"ja", "en", "fr", "de", "es"}
	for _, lang := range languages {
		if !strings.Contains(content, `"`+lang+`": []byte(`) {
			t.Errorf("Language %s not found in messageData", lang)
		}
	}
	
	t.Log("✓ Multi-language support test passed")
}

// Helper functions

func setupTestFiles(t *testing.T, tempDir, backend string, messages map[string]string) {
	// Create directories
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		t.Fatalf("Failed to create messages dir: %v", err)
	}
	if err := os.MkdirAll(placeholdersDir, 0755); err != nil {
		t.Fatalf("Failed to create placeholders dir: %v", err)
	}
	
	// Create config
	configContent := `
backend: ` + backend + `
compound: true
locales: [ja, en]
messages: "messages/*.yaml"
placeholders: "placeholders/*.yaml"
output_dir: "."
output_package: "i18n"
`
	
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	
	// Create message file
	messageContent := ""
	for msgName, msgTemplates := range messages {
		messageContent += msgName + ":" + msgTemplates + "\n"
	}
	
	messagePath := filepath.Join(messagesDir, "messages.yaml")
	if err := os.WriteFile(messagePath, []byte(messageContent), 0644); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}
	
	// Create minimal placeholder file
	placeholderContent := `
user:
  ja: ユーザー
  en: User
`
	
	placeholderPath := filepath.Join(placeholdersDir, "name.yaml")
	if err := os.WriteFile(placeholderPath, []byte(placeholderContent), 0644); err != nil {
		t.Fatalf("Failed to write placeholders: %v", err)
	}
}

func runGeneration(t *testing.T, tempDir string) error {
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
	configPath := filepath.Join(tempDir, "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}
	
	return generator.Run(cfg)
}

func readGeneratedFile(t *testing.T, tempDir string) string {
	outputPath := filepath.Join(tempDir, "i18n.gen.go")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	return string(content)
}