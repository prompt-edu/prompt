package slack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	providerpkg "github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

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

func newSlackProvider(t *testing.T, handler http.HandlerFunc) *Provider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	provider := New(Config{BotToken: "xoxb-test"})
	provider.client = server.Client()
	provider.baseURL = server.URL
	return provider
}

// Slack documents conversations.list and users.lookupByEmail as form-encoded; a JSON
// body is rejected with invalid_arguments.
func TestSlackSendsFormEncodedParameters(t *testing.T) {
	var contentType, name string
	provider := newSlackProvider(t, func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		_ = r.ParseForm()
		name = r.PostForm.Get("name")
		_, _ = w.Write([]byte(`{"ok":true,"channel":{"id":"C1"}}`))
	})

	if _, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "Team A"}); err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		t.Fatalf("content type = %q, want form encoding", contentType)
	}
	if name != "team-a" {
		t.Fatalf("channel name = %q, want team-a", name)
	}
}

// A page without response_metadata must end the loop. Reusing one response struct kept
// the previous cursor and re-requested the same page until the context expired.
func TestSlackPaginationStopsWithoutCursor(t *testing.T) {
	listCalls := 0
	provider := newSlackProvider(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "conversations.create"):
			_, _ = w.Write([]byte(`{"ok":false,"error":"name_taken"}`))
		case strings.HasSuffix(r.URL.Path, "conversations.list"):
			listCalls++
			if listCalls > 3 {
				t.Errorf("conversations.list called %d times: the cursor is not advancing", listCalls)
				_, _ = w.Write([]byte(`{"ok":true,"channels":[]}`))
				return
			}
			if listCalls == 1 {
				_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"C9","name":"other"}],"response_metadata":{"next_cursor":"page2"}}`))
				return
			}
			// Second page carries no response_metadata at all.
			_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"C2","name":"team-a"}]}`))
		default:
			t.Errorf("unexpected call: %s", r.URL.Path)
		}
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "Team A"})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "C2" {
		t.Fatalf("externalID = %q, want C2 from the second page", resource.ExternalID)
	}
	if listCalls != 2 {
		t.Fatalf("conversations.list calls = %d, want 2", listCalls)
	}
}

// An archived channel still owns the name, so it must be reported rather than looked
// past into a permanent "channel not found".
func TestSlackReportsArchivedChannelHoldingTheName(t *testing.T) {
	provider := newSlackProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "conversations.create") {
			_, _ = w.Write([]byte(`{"ok":false,"error":"name_taken"}`))
			return
		}
		_ = r.ParseForm()
		if r.PostForm.Get("exclude_archived") != "false" {
			t.Errorf("exclude_archived = %q, want false so the archived channel is visible", r.PostForm.Get("exclude_archived"))
		}
		_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"C3","name":"team-a","is_archived":true}]}`))
	})

	_, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "Team A"})
	if err == nil || !strings.Contains(err.Error(), "archived") {
		t.Fatalf("error = %v, want it to name the archived channel", err)
	}
}

// A member that cannot be resolved is a warning, not a silent skip.
func TestSlackReportsUnresolvedMemberAsWarning(t *testing.T) {
	provider := newSlackProvider(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "conversations.create"):
			_, _ = w.Write([]byte(`{"ok":true,"channel":{"id":"C1"}}`))
		case strings.HasSuffix(r.URL.Path, "users.lookupByEmail"):
			_, _ = w.Write([]byte(`{"ok":false,"error":"users_not_found"}`))
		default:
			t.Errorf("unexpected call: %s", r.URL.Path)
		}
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:    "Team A",
		Members: []providerpkg.Member{{Email: "ghost@example.com", Role: "student"}},
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if len(resource.Warnings) != 1 {
		t.Fatalf("warnings = %v, want one for the unresolved member", resource.Warnings)
	}
}

func TestSlackRejectsNameThatSanitizesToEmpty(t *testing.T) {
	provider := newSlackProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected for an empty channel name, got %s", r.URL.Path)
	})
	if _, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "---"}); err == nil {
		t.Fatal("CreateResource returned no error for a name that sanitizes to empty")
	}
}
