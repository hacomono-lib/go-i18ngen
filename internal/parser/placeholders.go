package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hacomono-lib/go-i18ngen/internal/model"

	"gopkg.in/yaml.v3"
)

// Pre-compiled regular expressions for better performance
var (
	identifierStartPattern = regexp.MustCompile(`^[a-zA-Z_]`)
	identifierPattern      = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// isValidGoIdentifier checks if a string is a valid Go identifier
func isValidGoIdentifier(name string) bool {
	if name == "" {
		return false
	}

	// Must start with letter or underscore
	if !identifierStartPattern.MatchString(name) {
		return false
	}

	// Must only contain letters, digits, and underscores
	if !identifierPattern.MatchString(name) {
		return false
	}

	return true
}

func ParsePlaceholders(pattern string, locales []string, compound bool) ([]model.PlaceholderSource, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern for placeholders %q: %w", pattern, err)
	}

	if len(files) == 0 {
		// Placeholders are optional, so return empty slice instead of error
		return []model.PlaceholderSource{}, nil
	}

	kindMap := map[string]map[string]map[string]string{} // kind -> id -> locale -> value

	for _, file := range files {
		base := filepath.Base(file)
		kind := strings.Split(base, ".")[0]
		ext := filepath.Ext(file)

		f, err := os.Open(file) // #nosec G304 - Opening placeholder files is intentional
		if err != nil {
			return nil, fmt.Errorf("failed to open placeholder file %q: %w", file, err)
		}
		defer func() { _ = f.Close() }()

		var parsed map[string]map[string]string
		if compound {
			parsed, err = decodeCompoundFile(f, ext)
			if err != nil {
				return nil, fmt.Errorf("failed to parse compound placeholder file %q (ext: %s): %w", file, ext, err)
			}
		} else {
			simple, err := decodeSimpleFile(f, ext)
			if err != nil {
				return nil, fmt.Errorf("failed to parse simple placeholder file %q (ext: %s, locale: %s): %w", file, ext, detectLocale(base), err)
			}
			parsed = make(map[string]map[string]string)
			for k, v := range simple {
				parsed[k] = map[string]string{detectLocale(base): v}
			}
		}

		if _, ok := kindMap[kind]; !ok {
			kindMap[kind] = map[string]map[string]string{}
		}

		for id, locMap := range parsed {
			if _, ok := kindMap[kind][id]; !ok {
				kindMap[kind][id] = map[string]string{}
			}
			for locale, val := range locMap {
				kindMap[kind][id][locale] = val
			}
		}
	}

	var results []model.PlaceholderSource
	for kind, items := range kindMap {
		// Validate placeholder kind name
		if !isValidGoIdentifier(kind) {
			return nil, fmt.Errorf("invalid placeholder kind name %q: must be a valid Go identifier (pattern: ^[a-zA-Z_][a-zA-Z0-9_]*$)", kind)
		}

		// Validate placeholder item IDs
		for id := range items {
			if !isValidGoIdentifier(id) {
				return nil, fmt.Errorf(
					"invalid placeholder item ID %q in kind %q: must be a valid Go identifier "+
						"(pattern: ^[a-zA-Z_][a-zA-Z0-9_]*$)", id, kind)
			}
		}

		results = append(results, model.PlaceholderSource{
			Kind:  kind,
			Items: items,
		})
	}
	return results, nil
}

func detectLocale(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

func decodeCompoundFile(file *os.File, ext string) (map[string]map[string]string, error) {
	var data map[string]map[string]string
	if ext == jsonExt {
		err := json.NewDecoder(file).Decode(&data)
		return data, err
	}
	err := yaml.NewDecoder(file).Decode(&data)
	return data, err
}

func decodeSimpleFile(file *os.File, ext string) (map[string]string, error) {
	var data map[string]string
	if ext == jsonExt {
		err := json.NewDecoder(file).Decode(&data)
		return data, err
	}
	err := yaml.NewDecoder(file).Decode(&data)
	return data, err
}
