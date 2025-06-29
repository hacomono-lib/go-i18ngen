package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/generator"
)

func TestMultiLanguageSupport(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	
	// Create test config for multiple languages
	configContent := `
backend: go-i18n
compound: true
locales: [ja, en, fr, de, es]
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
	
	// Create multi-language message file
	messageContent := `
WelcomeMessage:
  ja: "ようこそ、{{.name}}さん"
  en: "Welcome, {{.name}}"
  fr: "Bienvenue, {{.name}}"
  de: "Willkommen, {{.name}}"
  es: "Bienvenido, {{.name}}"

GoodbyeMessage:
  ja: "さようなら、{{.name}}さん"
  en: "Goodbye, {{.name}}"
  fr: "Au revoir, {{.name}}"
  de: "Auf Wiedersehen, {{.name}}"
  es: "Adiós, {{.name}}"
`
	
	messagePath := filepath.Join(messagesDir, "messages.yaml")
	if err := os.WriteFile(messagePath, []byte(messageContent), 0644); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}
	
	// Create placeholder file
	placeholderContent := `
user:
  ja: ユーザー
  en: User
  fr: Utilisateur
  de: Benutzer
  es: Usuario
`
	
	placeholderPath := filepath.Join(placeholdersDir, "user.yaml")
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
	
	// Check that all languages are included in messageData
	languages := []string{"ja", "en", "fr", "de", "es"}
	for _, lang := range languages {
		if !contains(generatedContent, `"`+lang+`": []byte(`) {
			t.Errorf("Language %s not found in messageData", lang)
		}
	}
	
	// Check that all messages are included
	messages := []string{"WelcomeMessage", "GoodbyeMessage"}
	for _, msg := range messages {
		if !contains(generatedContent, msg+`:`) {
			t.Errorf("Message %s not found in generated content", msg)
		}
	}
	
	// Check that placeholder data includes all languages
	for _, lang := range languages {
		if !contains(generatedContent, `"`+lang+`": "`) {
			t.Errorf("Language %s not found in placeholderData", lang)
		}
	}
	
	t.Logf("Multi-language test passed for languages: %v", languages)
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		   len(s) >= len(substr) && 
		   findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}