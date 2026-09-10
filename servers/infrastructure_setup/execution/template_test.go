package execution

import (
	"testing"
)

func TestResolveName(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		data     TemplateData
		expected string
	}{
		{
			name:     "lowercase student placeholders",
			tmpl:     "{{studentFirstName}}-{{studentLastName}}-{{studentEmail}}",
			data:     TemplateData{StudentFirstName: "Max", StudentLastName: "Muster", StudentEmail: "max@example.com"},
			expected: "max-muster-maxexample.com",
		},
		{
			name:     "lowercase team and semester placeholders",
			tmpl:     "{{teamName}}-{{semesterTag}}",
			data:     TemplateData{TeamName: "Team Alpha", SemesterTag: "ios26"},
			expected: "team-alpha-ios26",
		},
		{
			name:     "dotted spellings resolve to the same values",
			tmpl:     "{{.TeamName}}-{{.StudentLogin}}",
			data:     TemplateData{TeamName: "Team Alpha", StudentLogin: "ga12abc"},
			expected: "team-alpha-ga12abc",
		},
		{
			// Saved templates may use either spelling of the semester tag.
			name:     "semester aliases",
			tmpl:     "{{.Semester}}-{{.SemesterTag}}-{{semesterTag}}",
			data:     TemplateData{SemesterTag: "ios26"},
			expected: "ios26-ios26-ios26",
		},
		{
			name:     "special characters stripped",
			tmpl:     "{{.TeamName}}-channel",
			data:     TemplateData{TeamName: "Team (Alpha) #1"},
			expected: "team-alpha-1-channel",
		},
		{
			name:     "known placeholder with no value renders empty",
			tmpl:     "team-{{teamName}}",
			data:     TemplateData{},
			expected: "team-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveName(tt.tmpl, tt.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ResolveName(%q) = %q, want %q", tt.tmpl, got, tt.expected)
			}
		})
	}
}

// An unknown placeholder must fail rather than end up in the name of a real resource.
func TestResolveNameRejectsUnknownPlaceholder(t *testing.T) {
	for _, tmpl := range []string{"{{.CourseName}}-team", "team-{{.TeamIndex}}", "{{typo}}"} {
		if _, err := ResolveName(tmpl, TemplateData{TeamName: "Team A"}); err == nil {
			t.Errorf("ResolveName(%q) returned no error for an unknown placeholder", tmpl)
		}
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"foo_bar.baz", "foo_bar.baz"},
		{"Team (A)", "team-a"},
		{"  leading-trailing  ", "leading-trailing"},
		{"", ""},
		{"---", ""},
	}
	for _, tt := range tests {
		got := sanitize(tt.input)
		if got != tt.expected {
			t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
