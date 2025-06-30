package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CaseTestSuite struct {
	suite.Suite
}

func (s *CaseTestSuite) TestToCamelCase() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single word", "word", "Word"},
		{"snake_case", "user_name", "UserName"},
		{"multiple underscores", "first_name_last_name", "FirstNameLastName"},
		{"already camelCase", "UserName", "UserName"},
		{"with numbers", "user_123_name", "User123Name"},
		{"empty string", "", ""},
		{"single character", "a", "A"},
		{"underscore at start", "_field", "Field"},
		{"underscore at end", "field_", "Field"},
		{"multiple consecutive underscores", "field__name", "FieldName"},
		{"all underscores", "___", ""},
		{"mixed case with underscores", "First_Name_Last", "FirstNameLast"},
		{"long field name", "very_long_field_name_with_many_parts", "VeryLongFieldNameWithManyParts"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := ToCamelCase(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

// Run the test suite
func (s *CaseTestSuite) TestSafeGoIdentifier() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid identifier", "UserName", "UserName"},
		{"go keyword - type", "type", "type_"},
		{"go keyword - func", "func", "func_"},
		{"go keyword - var", "var", "var_"},
		{"go keyword - package", "package", "package_"},
		{"non-keyword", "user123_name", "user123_name"},
		{"mixed case keyword", "Type", "Type_"}, // "type" is a keyword, but this is case sensitive
		{"not a keyword", "types", "types"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := SafeGoIdentifier(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *CaseTestSuite) TestSafeGoIdentifierWithGoKeywords() {
	// Test all Go keywords
	keywords := []string{
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
	}

	for _, keyword := range keywords {
		s.Run("keyword_"+keyword, func() {
			result := SafeGoIdentifier(keyword)
			s.Equal(keyword+"_", result)
		})
	}
}

func (s *CaseTestSuite) TestToCamelCaseEdgeCases() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"mixed case input", "uSeR_nAmE", "USeRNAmE"},
		{"starts with number", "123_user", "123User"},
		{"complex case", "api_v2_user_profile", "ApiV2UserProfile"},
		{"consecutive separators", "user___name", "UserName"},
		{"ends with separator", "user_", "User"},
		{"starts with separator", "_user", "User"},
		{"empty parts", "user__name", "UserName"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := ToCamelCase(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func TestCaseSuite(t *testing.T) {
	suite.Run(t, new(CaseTestSuite))
}
