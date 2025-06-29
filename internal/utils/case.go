// Package utils provides utility functions for string manipulation and formatting.
package utils

import "strings"

// ToCamelCase converts snake_case to CamelCase (e.g. user_name -> UserName)
func ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if parts[i] != "" {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// Go reserved words that cannot be used as identifiers
var goReservedWords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// SafeGoIdentifier escapes Go reserved words by appending underscore
func SafeGoIdentifier(name string) string {
	if goReservedWords[strings.ToLower(name)] {
		return name + "_"
	}
	return name
}
