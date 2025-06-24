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
