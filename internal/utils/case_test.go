package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCamelCase(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := ToCamelCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
