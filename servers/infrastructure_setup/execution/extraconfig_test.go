package execution

import (
	"strings"
	"testing"
)

func TestResolveTemplatedExtraConfigOnlyTouchesDeclaredKeys(t *testing.T) {
	extra := map[string]interface{}{
		"parent_group_template": "{{semesterTag}}-{{teamName}}",
		"roleTemplateId":        "project-member-{{notAPlaceholder}}",
	}
	data := TemplateData{TeamName: "Team 1", SemesterTag: "ios2526"}

	resolved, err := ResolveTemplatedExtraConfig(extra, []string{"parent_group_template"}, data)
	if err != nil {
		t.Fatalf("ResolveTemplatedExtraConfig: %v", err)
	}
	if resolved["parent_group_template"] != "ios2526-team-1" {
		t.Fatalf("parent_group_template = %v, want the resolved name", resolved["parent_group_template"])
	}
	// An undeclared key keeps its braces: it is a provider setting, not a template.
	if resolved["roleTemplateId"] != "project-member-{{notAPlaceholder}}" {
		t.Fatalf("roleTemplateId = %v, want the literal value", resolved["roleTemplateId"])
	}
}

func TestResolveTemplatedExtraConfigRejectsUnknownPlaceholder(t *testing.T) {
	extra := map[string]interface{}{"parent_group_template": "{{nope}}"}

	_, err := ResolveTemplatedExtraConfig(extra, []string{"parent_group_template"}, TemplateData{})
	if err == nil {
		t.Fatal("ResolveTemplatedExtraConfig = nil error, want a rejection")
	}
	if !strings.Contains(err.Error(), "parent_group_template") {
		t.Fatalf("error = %v, want it to name the offending key", err)
	}
}

// The input must not mutate the caller's map: the worker reuses the parsed config.
func TestResolveTemplatedExtraConfigDoesNotMutateInput(t *testing.T) {
	extra := map[string]interface{}{"parent_group_template": "{{teamName}}"}

	if _, err := ResolveTemplatedExtraConfig(extra, []string{"parent_group_template"}, TemplateData{TeamName: "Team 1"}); err != nil {
		t.Fatalf("ResolveTemplatedExtraConfig: %v", err)
	}
	if extra["parent_group_template"] != "{{teamName}}" {
		t.Fatalf("input was mutated: %v", extra["parent_group_template"])
	}
}
