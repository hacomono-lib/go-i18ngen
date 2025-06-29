package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/generator"
)

func TestBuiltinCompatibility(t *testing.T) {
	// Test that builtin backend still works exactly as before
	tempDir := t.TempDir()
	
	// Create config for builtin backend
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
	
	// Create config file
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	
	// Create message file with template functions
	messageContent := `
EntityNotFound:
  ja: "{{.entity}}が見つかりません: {{.reason}}"
  en: "{{.entity}} not found: {{.reason}}"

ValidationError:
  ja: "{{.field:input}}の{{.field:display | upper}}検証エラーです"
  en: "{{.field:input | title}} validation error for {{.field:display}}"
`
	
	messagePath := filepath.Join(messagesDir, "messages.yaml")
	if err := os.WriteFile(messagePath, []byte(messageContent), 0644); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}
	
	// Create placeholder files
	entityContent := `
user:
  ja: ユーザー
  en: User
product:
  ja: 製品
  en: Product
`
	
	entityPath := filepath.Join(placeholdersDir, "entity.yaml")
	if err := os.WriteFile(entityPath, []byte(entityContent), 0644); err != nil {
		t.Fatalf("Failed to write entity placeholders: %v", err)
	}
	
	reasonContent := `
already_deleted:
  ja: すでに削除されています
  en: already deleted
not_found:
  ja: 見つかりません
  en: not found
`
	
	reasonPath := filepath.Join(placeholdersDir, "reason.yaml")
	if err := os.WriteFile(reasonPath, []byte(reasonContent), 0644); err != nil {
		t.Fatalf("Failed to write reason placeholders: %v", err)
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
	
	// Verify builtin-specific features
	t.Run("builtin imports", func(t *testing.T) {
		// Should not import go-i18n
		if strings.Contains(generatedContent, "github.com/nicksnyder/go-i18n") {
			t.Error("Builtin backend should not import go-i18n")
		}
		
		// Should import text/template
		if !strings.Contains(generatedContent, `"text/template"`) {
			t.Error("Builtin backend should import text/template")
		}
	})
	
	t.Run("builtin template system", func(t *testing.T) {
		// Should have template cache
		if !strings.Contains(generatedContent, "templateCache") {
			t.Error("Builtin backend should have templateCache")
		}
		
		// Should have renderTemplate function
		if !strings.Contains(generatedContent, "func renderTemplate") {
			t.Error("Builtin backend should have renderTemplate function")
		}
		
		// Should not have bundle/localizer
		if strings.Contains(generatedContent, "bundle") || strings.Contains(generatedContent, "localizer") {
			t.Error("Builtin backend should not have bundle/localizer")
		}
	})
	
	t.Run("template functions", func(t *testing.T) {
		// Should have getTemplateFunctions
		if !strings.Contains(generatedContent, "getTemplateFunctions") {
			t.Error("Builtin backend should have getTemplateFunctions")
		}
		
		// Should support title, upper, lower
		if !strings.Contains(generatedContent, `"title"`) ||
		   !strings.Contains(generatedContent, `"upper"`) ||
		   !strings.Contains(generatedContent, `"lower"`) {
			t.Error("Builtin backend should support title, upper, lower functions")
		}
	})
	
	t.Run("message structure", func(t *testing.T) {
		// Should have message templates map
		if !strings.Contains(generatedContent, "var templates = map[string]map[string]string{") {
			t.Error("Builtin backend should have templates map")
		}
		
		// Should have EntityNotFound struct
		if !strings.Contains(generatedContent, "type EntityNotFound struct") {
			t.Error("Should generate EntityNotFound struct")
		}
		
		// Should have NewEntityNotFound function
		if !strings.Contains(generatedContent, "func NewEntityNotFound") {
			t.Error("Should generate NewEntityNotFound function")
		}
	})
	
	t.Run("placeholder structure", func(t *testing.T) {
		// Should have EntityTexts utility
		if !strings.Contains(generatedContent, "var EntityTexts = struct") {
			t.Error("Should generate EntityTexts utility")
		}
		
		// Should have ReasonTexts utility
		if !strings.Contains(generatedContent, "var ReasonTexts = struct") {
			t.Error("Should generate ReasonTexts utility")
		}
	})
	
	t.Logf("Builtin compatibility test passed")
}

func TestBackendComparison(t *testing.T) {
	// Test that both backends generate compatible APIs
	tempDir := t.TempDir()
	
	// Common test data
	messageContent := `
SimpleMessage:
  ja: "こんにちは、{{.name}}さん"
  en: "Hello, {{.name}}"
`
	
	placeholderContent := `
admin:
  ja: 管理者
  en: Admin
`
	
	// Test both backends
	backends := []string{"builtin", "go-i18n"}
	
	for _, backend := range backends {
		t.Run("backend_"+backend, func(t *testing.T) {
			backendDir := filepath.Join(tempDir, backend)
			messagesDir := filepath.Join(backendDir, "messages")
			placeholdersDir := filepath.Join(backendDir, "placeholders")
			
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
			
			configPath := filepath.Join(backendDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}
			
			messagePath := filepath.Join(messagesDir, "messages.yaml")
			if err := os.WriteFile(messagePath, []byte(messageContent), 0644); err != nil {
				t.Fatalf("Failed to write messages: %v", err)
			}
			
			placeholderPath := filepath.Join(placeholdersDir, "name.yaml")
			if err := os.WriteFile(placeholderPath, []byte(placeholderContent), 0644); err != nil {
				t.Fatalf("Failed to write placeholders: %v", err)
			}
			
			// Change to backend directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			defer os.Chdir(originalDir)
			
			if err := os.Chdir(backendDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}
			
			// Generate
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}
			
			if err := generator.Run(cfg); err != nil {
				t.Fatalf("Generation failed for %s backend: %v", backend, err)
			}
			
			// Verify compatible API
			outputPath := filepath.Join(backendDir, "i18n.gen.go")
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}
			
			generatedContent := string(content)
			
			// Both backends should generate the same API
			if !strings.Contains(generatedContent, "type SimpleMessage struct") {
				t.Errorf("%s backend should generate SimpleMessage struct", backend)
			}
			
			if !strings.Contains(generatedContent, "func NewSimpleMessage") {
				t.Errorf("%s backend should generate NewSimpleMessage function", backend)
			}
			
			if !strings.Contains(generatedContent, "func (m SimpleMessage) Localize(locale string) string") {
				t.Errorf("%s backend should generate Localize method", backend)
			}
			
			if !strings.Contains(generatedContent, "var NameTexts = struct") {
				t.Errorf("%s backend should generate NameTexts utility", backend)
			}
		})
	}
}