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
			name:     "team name template",
			tmpl:     "{{.CourseName}}-team-{{.TeamIndex}}",
			data:     TemplateData{CourseName: "iPraktikum", TeamIndex: 3},
			expected: "ipraktikum-team-3",
		},
		{
			name:     "student template",
			tmpl:     "{{.CourseName}}-{{.StudentLogin}}",
			data:     TemplateData{CourseName: "iPraktikum", StudentLogin: "ga12abc"},
			expected: "ipraktikum-ga12abc",
		},
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
			name:     "spaces in course name become hyphens",
			tmpl:     "{{.CourseName}}-group",
			data:     TemplateData{CourseName: "My Course"},
			expected: "my-course-group",
		},
		{
			name:     "special characters stripped",
			tmpl:     "{{.TeamName}}-channel",
			data:     TemplateData{TeamName: "Team (Alpha) #1"},
			expected: "team-alpha-1-channel",
		},
		{
			name:     "semester included",
			tmpl:     "{{.CourseName}}-{{.Semester}}{{.Year}}",
			data:     TemplateData{CourseName: "iPraktikum", Semester: "WS", Year: "2024"},
			expected: "ipraktikum-ws2024",
		},
		{
			name:     "empty placeholder leaves no artifacts",
			tmpl:     "{{.CourseName}}-{{.TeamName}}",
			data:     TemplateData{CourseName: "Course"},
			expected: "course-",
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
