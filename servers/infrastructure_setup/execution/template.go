// Package execution handles the lifecycle of resource instances:
// resolving names from templates, spawning workers, and reporting results.
package execution

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// TemplateData holds the variables available for name template substitution.
type TemplateData struct {
	CourseName  string
	TeamName    string
	TeamIndex   int
	StudentName string
	StudentLogin string
	Semester    string
	Year        string
}

// ResolveName resolves a name template string against the provided data.
// Supported placeholders: {{.CourseName}}, {{.TeamName}}, {{.TeamIndex}},
// {{.StudentName}}, {{.StudentLogin}}, {{.Semester}}, {{.Year}}.
// Braces-style (not Go text/template) keeps the server dependency-light.
func ResolveName(tmpl string, data TemplateData) (string, error) {
	replacements := map[string]string{
		"{{.CourseName}}":   sanitize(data.CourseName),
		"{{.TeamName}}":     sanitize(data.TeamName),
		"{{.TeamIndex}}":    fmt.Sprintf("%d", data.TeamIndex),
		"{{.StudentName}}":  sanitize(data.StudentName),
		"{{.StudentLogin}}": sanitize(data.StudentLogin),
		"{{.Semester}}":     sanitize(data.Semester),
		"{{.Year}}":         sanitize(data.Year),
	}

	result := tmpl
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result, nil
}

// sanitize strips characters that are problematic in resource names across providers.
// Keeps alphanumeric, hyphens, underscores, and dots. Replaces spaces with hyphens.
func sanitize(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' {
			b.WriteRune(unicode.ToLower(r))
			prevHyphen = false
		} else if (r == ' ' || r == '-') && !prevHyphen {
			b.WriteRune('-')
			prevHyphen = true
		}
	}
	return strings.Trim(b.String(), "-_.")
}

// ParsePermissionMapping unmarshals the JSONB permission mapping from the DB.
// Returns an empty map on nil/empty input.
func ParsePermissionMapping(raw json.RawMessage) (map[string]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]string{}, nil
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parsing permission mapping: %w", err)
	}
	return m, nil
}

// ParseExtraConfig unmarshals the JSONB extra config from the DB.
// Returns an empty map on nil/empty input.
func ParseExtraConfig(raw json.RawMessage) (map[string]interface{}, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]interface{}{}, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parsing extra config: %w", err)
	}
	return m, nil
}
