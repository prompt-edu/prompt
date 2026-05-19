package gitlab

import "testing"

func TestGitLabNameNormalization(t *testing.T) {
	if got := sanitizeName("Team (Alpha) #1"); got != "Team -Alpha- -1" {
		t.Fatalf("sanitizeName = %q, want %q", got, "Team -Alpha- -1")
	}
	if got := toSlug("Team (Alpha) #1"); got != "team-alpha-1" {
		t.Fatalf("toSlug = %q, want team-alpha-1", got)
	}
}

func TestGitLabAccessLevel(t *testing.T) {
	tests := map[string]int{
		"guest":      10,
		"reporter":   20,
		"developer":  30,
		"maintainer": 40,
		"owner":      50,
		"unknown":    10,
	}
	for permission, expected := range tests {
		if got := gitlabAccessLevel(permission); got != expected {
			t.Fatalf("gitlabAccessLevel(%q) = %d, want %d", permission, got, expected)
		}
	}
}

func TestGitLabAuthFields(t *testing.T) {
	fields := (&Provider{}).GetAuthFields()
	if len(fields) != 3 {
		t.Fatalf("auth fields = %d, want 3", len(fields))
	}
	if fields[0].Name != "base_url" || !fields[0].Required {
		t.Fatalf("first auth field = %+v, want required base_url", fields[0])
	}
}
