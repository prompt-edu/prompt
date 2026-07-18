package slack

import "testing"

func TestSanitizeChannelName(t *testing.T) {
	tests := map[string]string{
		"Team Alpha":      "team-alpha",
		"Team (Alpha) #1": "team-alpha-1",
		"---":             "",
	}
	for input, expected := range tests {
		if got := sanitizeChannelName(input); got != expected {
			t.Fatalf("sanitizeChannelName(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestSanitizeChannelNameTruncatesToSlackLimit(t *testing.T) {
	input := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	got := sanitizeChannelName(input)
	if len(got) != 80 {
		t.Fatalf("sanitized channel length = %d, want 80", len(got))
	}
}

func TestSlackAuthFields(t *testing.T) {
	fields := (&Provider{}).GetAuthFields()
	if len(fields) != 1 {
		t.Fatalf("auth fields = %d, want 1", len(fields))
	}
	if fields[0].Name != "bot_token" || fields[0].Type != "password" || !fields[0].Required {
		t.Fatalf("auth field = %+v, want required bot_token password", fields[0])
	}
}
