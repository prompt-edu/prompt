// Package execution handles the lifecycle of resource instances:
// resolving names from templates, spawning workers, and reporting results.
package execution

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// TemplateData holds the variables available for name template substitution.
type TemplateData struct {
	TeamName         string
	StudentName      string
	StudentFirstName string
	StudentLastName  string
	StudentEmail     string
	StudentLogin     string
	SemesterTag      string
}

// unresolvedPlaceholder matches any {{...}} left after substitution.
var unresolvedPlaceholder = regexp.MustCompile(`{{[^}]*}}`)

// ResolveName resolves a name template string against the provided data.
//
// Braces-style (not Go text/template) keeps the server dependency-light. Both the
// dotted and the lowercase spelling of each placeholder are accepted because saved
// templates use either. An unresolved placeholder is an error rather than an empty
// string: it would otherwise end up in the name of a real external resource.
func ResolveName(tmpl string, data TemplateData) (string, error) {
	replacements := map[string]string{
		"{{.TeamName}}":        sanitize(data.TeamName),
		"{{teamName}}":         sanitize(data.TeamName),
		"{{.StudentName}}":     sanitize(data.StudentName),
		"{{studentName}}":      sanitize(data.StudentName),
		"{{studentFirstName}}": sanitize(data.StudentFirstName),
		"{{studentLastName}}":  sanitize(data.StudentLastName),
		"{{studentEmail}}":     sanitize(data.StudentEmail),
		"{{.StudentLogin}}":    sanitize(data.StudentLogin),
		"{{studentLogin}}":     sanitize(data.StudentLogin),
		"{{semesterTag}}":      sanitize(data.SemesterTag),
		"{{.SemesterTag}}":     sanitize(data.SemesterTag),
		"{{.Semester}}":        sanitize(data.SemesterTag),
	}

	result := tmpl
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	if leftover := unresolvedPlaceholder.FindString(result); leftover != "" {
		return "", fmt.Errorf("unknown placeholder %s in name template", leftover)
	}
	return result, nil
}

// SupportedPlaceholders lists the placeholders a name template may use.
func SupportedPlaceholders() []string {
	return []string{
		"{{teamName}}",
		"{{studentName}}",
		"{{studentFirstName}}",
		"{{studentLastName}}",
		"{{studentEmail}}",
		"{{studentLogin}}",
		"{{semesterTag}}",
	}
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

// ResolveTemplatedExtraConfig resolves placeholders in the extra-config values a
// provider declares as templates, leaving every other value literal so an ordinary
// provider setting can still contain braces.
func ResolveTemplatedExtraConfig(extra map[string]interface{}, keys []string, data TemplateData) (map[string]interface{}, error) {
	if len(extra) == 0 || len(keys) == 0 {
		return extra, nil
	}

	resolved := make(map[string]interface{}, len(extra))
	for key, value := range extra {
		resolved[key] = value
	}

	for _, key := range keys {
		raw, ok := resolved[key]
		if !ok {
			continue
		}
		text, isString := raw.(string)
		if !isString {
			return nil, fmt.Errorf("extra config %q must be a string", key)
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		value, err := ResolveName(text, data)
		if err != nil {
			return nil, fmt.Errorf("extra config %q: %w", key, err)
		}
		resolved[key] = value
	}
	return resolved, nil
}
