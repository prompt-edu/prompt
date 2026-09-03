// Package execution handles the lifecycle of resource instances:
// resolving names from templates, spawning workers, and reporting results.
package execution

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
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

// placeholderAvailability says which scopes can fill a placeholder.
type placeholderAvailability uint8

const (
	// availableAlways marks a placeholder every scope can fill.
	availableAlways placeholderAvailability = iota
	// availableForTeams marks a placeholder only per_team resolution fills.
	availableForTeams
	// availableForStudents marks a placeholder only per_student resolution fills.
	availableForStudents
)

// placeholder describes one substitutable value and every spelling of it a saved
// template may use. It is the single source of truth for resolution, for the list the
// UI offers and for what a scope is allowed to reference.
type placeholder struct {
	// canonical is the spelling reported to lecturers.
	canonical string
	// aliases are additional accepted spellings.
	aliases []string
	// availability says which scopes fill this placeholder.
	availability placeholderAvailability
	// value reads the placeholder's value out of the resolved target data.
	value func(TemplateData) string
}

func (p placeholder) tokens() []string {
	return append([]string{p.canonical}, p.aliases...)
}

var placeholders = []placeholder{
	{
		canonical:    "{{teamName}}",
		aliases:      []string{"{{.TeamName}}"},
		availability: availableForTeams,
		value:        func(d TemplateData) string { return d.TeamName },
	},
	{
		canonical:    "{{studentName}}",
		aliases:      []string{"{{.StudentName}}"},
		availability: availableForStudents,
		value:        func(d TemplateData) string { return d.StudentName },
	},
	{
		canonical:    "{{studentFirstName}}",
		availability: availableForStudents,
		value:        func(d TemplateData) string { return d.StudentFirstName },
	},
	{
		canonical:    "{{studentLastName}}",
		availability: availableForStudents,
		value:        func(d TemplateData) string { return d.StudentLastName },
	},
	{
		canonical:    "{{studentEmail}}",
		availability: availableForStudents,
		value:        func(d TemplateData) string { return d.StudentEmail },
	},
	{
		canonical:    "{{studentLogin}}",
		aliases:      []string{"{{.StudentLogin}}"},
		availability: availableForStudents,
		value:        func(d TemplateData) string { return d.StudentLogin },
	},
	{
		canonical:    "{{semesterTag}}",
		aliases:      []string{"{{.SemesterTag}}", "{{.Semester}}"},
		availability: availableAlways,
		value:        func(d TemplateData) string { return d.SemesterTag },
	},
}

// ResolveName resolves a name template string against the provided data.
//
// Braces-style (not Go text/template) keeps the server dependency-light. Both the
// dotted and the lowercase spelling of each placeholder are accepted because saved
// templates use either. An unresolved placeholder is an error rather than an empty
// string: it would otherwise end up in the name of a real external resource.
func ResolveName(tmpl string, data TemplateData) (string, error) {
	result := tmpl
	for _, p := range placeholders {
		value := sanitize(p.value(data))
		for _, token := range p.tokens() {
			result = strings.ReplaceAll(result, token, value)
		}
	}

	if leftover := unresolvedPlaceholder.FindString(result); leftover != "" {
		return "", fmt.Errorf("unknown placeholder %s in name template", leftover)
	}
	return result, nil
}

// SupportedPlaceholders lists the placeholders a name template may use.
func SupportedPlaceholders() []string {
	supported := make([]string, 0, len(placeholders))
	for _, p := range placeholders {
		supported = append(supported, p.canonical)
	}
	return supported
}

// PlaceholdersForScope lists the placeholders a scope's resolution actually fills.
func PlaceholdersForScope(scope db.ResourceScope) []string {
	available := make([]string, 0, len(placeholders))
	for _, p := range placeholders {
		if placeholderFilledBy(p, scope) {
			available = append(available, p.canonical)
		}
	}
	return available
}

// UnfillablePlaceholders lists the placeholders a template uses that the scope cannot
// fill. Resolution would replace each of them with an empty string, so every target of
// the config would converge on the same name - one shared external resource with every
// team's or student's members in it.
func UnfillablePlaceholders(tmpl string, scope db.ResourceScope) []string {
	var unfillable []string
	for _, p := range placeholders {
		if placeholderFilledBy(p, scope) {
			continue
		}
		for _, token := range p.tokens() {
			if strings.Contains(tmpl, token) {
				unfillable = append(unfillable, token)
			}
		}
	}
	return unfillable
}

func placeholderFilledBy(p placeholder, scope db.ResourceScope) bool {
	switch p.availability {
	case availableAlways:
		return true
	case availableForTeams:
		return scope == db.ResourceScopePerTeam
	case availableForStudents:
		return scope == db.ResourceScopePerStudent
	}
	return false
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
