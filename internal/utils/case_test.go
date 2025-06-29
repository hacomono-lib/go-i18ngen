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
func TestCaseSuite(t *testing.T) {
	suite.Run(t, new(CaseTestSuite))
}