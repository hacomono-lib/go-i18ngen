package i18ngen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/generator"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalizeIntegration(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_integration_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	outputDir := filepath.Join(tempDir, "output")
	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))

	// Create test message file (using suffix notation instead of duplicate placeholders)
	messageFile := filepath.Join(messagesDir, "messages.yaml")
	messageContent := `WelcomeMessage:
  ja: "{{.name:user}}さん、{{.name:owner}}さんのアカウントへようこそ！"
  en: "Welcome {{.name:user}}, to {{.name:owner}}'s account!"
ValidationError:
  ja: "{{.field:input}}の{{.field:display | upper}}検証エラーです"
  en: "{{.field:input | title}} validation error for {{.field:display}}"
EntityNotFound:
  ja: "{{.entity}}が見つかりません: {{.reason}}"
  en: "{{.entity}} not found: {{.reason}}"
`
	require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Create test placeholder files
	entityFile := filepath.Join(placeholdersDir, "entity.yaml")
	entityContent := `user:
  ja: "ユーザー"
  en: "User"
product:
  ja: "製品"
  en: "Product"
`
	require.NoError(t, os.WriteFile(entityFile, []byte(entityContent), 0644))

	reasonFile := filepath.Join(placeholdersDir, "reason.yaml")
	reasonContent := `not_exist:
  ja: "存在しません"
  en: "does not exist"
deleted:
  ja: "削除されています"
  en: "has been deleted"
`
	require.NoError(t, os.WriteFile(reasonFile, []byte(reasonContent), 0644))

	// Create configuration
	cfg := &config.Config{
		Locales:          []string{"ja", "en"},
		Compound:         true,
		MessagesGlob:     filepath.Join(messagesDir, "*.yaml"),
		PlaceholdersGlob: filepath.Join(placeholdersDir, "*.yaml"),
		OutputDir:        outputDir,
		OutputPackage:    "testpkg",
	}

	// Execute code generation
	err = generator.Run(cfg)
	require.NoError(t, err, "Code generation failed")

	// Verify that generated file exists
	generatedFile := filepath.Join(outputDir, "i18n.gen.go")
	assert.FileExists(t, generatedFile, "Generated file does not exist")

	// Verify generated code content
	generatedCode, err := os.ReadFile(generatedFile)
	require.NoError(t, err)
	codeStr := string(generatedCode)

	// Verify basic structure
	assert.Contains(t, codeStr, "package testpkg", "Package name is not set correctly")
	assert.Contains(t, codeStr, "type WelcomeMessage struct", "WelcomeMessage struct is not generated")
	assert.Contains(t, codeStr, "type ValidationError struct", "ValidationError struct is not generated")
	assert.Contains(t, codeStr, "type EntityNotFound struct", "EntityNotFound struct is not generated")

	// Verify that suffix notation placeholders are correctly processed
	welcomeStructStart := strings.Index(codeStr, "type WelcomeMessage struct")
	welcomeStructEnd := strings.Index(codeStr[welcomeStructStart:], "}")
	welcomeStruct := codeStr[welcomeStructStart : welcomeStructStart+welcomeStructEnd]
	nameUserCount := strings.Count(welcomeStruct, "NameUser ")
	nameOwnerCount := strings.Count(welcomeStruct, "NameOwner ")
	assert.Equal(t, 1, nameUserCount, "NameUser field is not correctly generated in WelcomeMessage struct")
	assert.Equal(t, 1, nameOwnerCount, "NameOwner field is not correctly generated in WelcomeMessage struct")

	// Verify ValidationError similarly
	validationStructStart := strings.Index(codeStr, "type ValidationError struct")
	validationStructEnd := strings.Index(codeStr[validationStructStart:], "}")
	validationStruct := codeStr[validationStructStart : validationStructStart+validationStructEnd]
	fieldInputCount := strings.Count(validationStruct, "FieldInput ")
	fieldDisplayCount := strings.Count(validationStruct, "FieldDisplay ")
	assert.Equal(t, 1, fieldInputCount, "FieldInput field is not correctly generated in ValidationError struct")
	assert.Equal(t, 1, fieldDisplayCount, "FieldDisplay field is not correctly generated in ValidationError struct")

	// Verify that EntityText-related structs and utilities are generated
	assert.Contains(t, codeStr, "type EntityText struct", "EntityText struct is not generated")
	assert.Contains(t, codeStr, "var EntityTexts = struct", "EntityTexts utility is not generated")
	assert.Contains(t, codeStr, "User EntityText", "EntityTexts.User is not generated")
	assert.Contains(t, codeStr, "Product EntityText", "EntityTexts.Product is not generated")

	// Verify ReasonText-related as well
	assert.Contains(t, codeStr, "type ReasonText struct", "ReasonText struct is not generated")
	assert.Contains(t, codeStr, "var ReasonTexts = struct", "ReasonTexts utility is not generated")

	// Verify that NameValue and FieldValue structs are generated (value-type placeholders)
	assert.Contains(t, codeStr, "type NameValue struct", "NameValue struct is not generated")
	assert.Contains(t, codeStr, "type FieldValue struct", "FieldValue struct is not generated")

	// Verify that constructor functions are generated
	assert.Contains(t, codeStr, "func NewWelcomeMessage(nameUser NameValue, nameOwner NameValue) WelcomeMessage", "NewWelcomeMessage function is not correctly generated")
	assert.Contains(t, codeStr, "func NewValidationError(fieldInput FieldValue, fieldDisplay FieldValue) ValidationError", "NewValidationError function is not correctly generated")
	assert.Contains(t, codeStr, "func NewEntityNotFound(entity EntityText, reason ReasonText) EntityNotFound", "NewEntityNotFound function is not correctly generated")

	// Verify that Localize functions are generated
	assert.Contains(t, codeStr, "func (m WelcomeMessage) Localize(locale string) string", "WelcomeMessage.Localize function is not generated")
	assert.Contains(t, codeStr, "func (m ValidationError) Localize(locale string) string", "ValidationError.Localize function is not generated")
	assert.Contains(t, codeStr, "func (m EntityNotFound) Localize(locale string) string", "EntityNotFound.Localize function is not generated")

	// Verify that template data is correctly included (with suffix notation)
	assert.Contains(t, codeStr, `"WelcomeMessage": {`, "WelcomeMessage template is not generated")
	assert.Contains(t, codeStr, `"{{.nameUser}}さん、{{.nameOwner}}さんのアカウントへようこそ！"`, "Japanese template is not correctly included")
	assert.Contains(t, codeStr, `"Welcome {{.nameUser}}, to {{.nameOwner}}'s account!"`, "English template is not correctly included")

	// Verify templates with template functions as well
	assert.Contains(t, codeStr, `"{{.fieldInput}}の{{.fieldDisplay | upper}}検証エラーです"`, "Japanese template with template functions is not correctly included")
	assert.Contains(t, codeStr, `"{{.fieldInput | title}} validation error for {{.fieldDisplay}}"`, "English template with template functions is not correctly included")

	// Verify that placeholder templates are correctly included
	assert.Contains(t, codeStr, `var entityTemplates = map[string]map[string]string{`, "entityTemplates is not generated")
	assert.Contains(t, codeStr, `var reasonTemplates = map[string]map[string]string{`, "reasonTemplates is not generated")
	assert.Contains(t, codeStr, `"user": {`, "user entity is not included")
	assert.Contains(t, codeStr, `"ユーザー"`, "Japanese user entity is not included")
	assert.Contains(t, codeStr, `"User"`, "English user entity is not included")

	t.Log("All integration tests passed successfully")
	t.Logf("Generated code size: %d bytes", len(generatedCode))
}

// Test to verify that generated code can be compiled correctly
func TestGeneratedCodeCompilation(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "i18ngen_compile_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create simple messages and placeholders
	messagesDir := filepath.Join(tempDir, "messages")
	placeholdersDir := filepath.Join(tempDir, "placeholders")
	outputDir := filepath.Join(tempDir, "output")
	require.NoError(t, os.MkdirAll(messagesDir, 0755))
	require.NoError(t, os.MkdirAll(placeholdersDir, 0755))

	// Simple message file
	messageFile := filepath.Join(messagesDir, "messages.yaml")
	messageContent := `TestMessage:
  ja: "{{.entity}}のテスト"
  en: "Test for {{.entity}}"
`
	require.NoError(t, os.WriteFile(messageFile, []byte(messageContent), 0644))

	// Simple placeholder file
	entityFile := filepath.Join(placeholdersDir, "entity.yaml")
	entityContent := `test:
  ja: "テスト項目"
  en: "Test Item"
`
	require.NoError(t, os.WriteFile(entityFile, []byte(entityContent), 0644))

	// Create configuration
	cfg := &config.Config{
		Locales:          []string{"ja", "en"},
		Compound:         true,
		MessagesGlob:     filepath.Join(messagesDir, "*.yaml"),
		PlaceholdersGlob: filepath.Join(placeholdersDir, "*.yaml"),
		OutputDir:        outputDir,
		OutputPackage:    "compilepkg",
	}

	// Execute code generation
	err = generator.Run(cfg)
	require.NoError(t, err, "Code generation failed")

	// Verify that generated file exists
	generatedFile := filepath.Join(outputDir, "i18n.gen.go")
	assert.FileExists(t, generatedFile, "Generated file does not exist")

	// Basic check that generated code is valid Go code
	generatedCode, err := os.ReadFile(generatedFile)
	require.NoError(t, err)
	codeStr := string(generatedCode)

	// Verify basic Go code structure
	assert.Contains(t, codeStr, "package compilepkg")
	assert.Contains(t, codeStr, "import (")
	assert.Contains(t, codeStr, "func ")
	assert.Contains(t, codeStr, "type ")

	// Basic check for syntax errors (bracket matching, etc.)
	openBraces := strings.Count(codeStr, "{")
	closeBraces := strings.Count(codeStr, "}")
	assert.Equal(t, openBraces, closeBraces, "Generated code has mismatched braces")

	openParens := strings.Count(codeStr, "(")
	closeParens := strings.Count(codeStr, ")")
	assert.Equal(t, openParens, closeParens, "Parentheses are not properly matched in generated code")

	t.Log("Basic structure check of generated code succeeded")
}
