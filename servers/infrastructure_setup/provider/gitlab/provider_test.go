package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	providerpkg "github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

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

// requestRecorder captures the paths a provider calls so a test can assert which
// GitLab endpoints were used.
type requestRecorder struct {
	paths []string
}

func newGitLabServer(t *testing.T, rec *requestRecorder, handler func(w http.ResponseWriter, r *http.Request)) *Provider {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.paths = append(rec.paths, r.URL.Path+"?"+r.URL.RawQuery)
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	return New(Config{BaseURL: server.URL, PrivateToken: "token"})
}

// A group under a different parent must not be adopted: matching on a full_path
// suffix would have added the students to another course's group.
func TestGitLabDoesNotAdoptGroupUnderAnotherParent(t *testing.T) {
	parentID := 42
	rec := &requestRecorder{}
	provider := newGitLabServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v4/groups/42/subgroups"):
			// The parent holds no matching subgroup.
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v4/groups":
			_, _ = w.Write([]byte(`{"id":7,"web_url":"https://gitlab.test/groups/my-course/team-a"}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
	})
	provider.cfg.ParentGroupID = &parentID

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:         "Team A",
		ResourceType: "group",
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "7" {
		t.Fatalf("externalID = %q, want 7 (a new group under the configured parent)", resource.ExternalID)
	}
	// The search must be scoped to the parent, not the whole instance.
	if !strings.Contains(rec.paths[0], "/api/v4/groups/42/subgroups") {
		t.Fatalf("first request = %q, want a subgroups lookup on the parent", rec.paths[0])
	}
}

// An existing subgroup of the configured parent is reused rather than recreated.
func TestGitLabReusesExistingSubgroup(t *testing.T) {
	parentID := 42
	rec := &requestRecorder{}
	provider := newGitLabServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.Errorf("unexpected create call: %s", r.URL)
		}
		_, _ = w.Write([]byte(`[{"id":9,"path":"team-a","full_path":"my-course/team-a","web_url":"https://gitlab.test/groups/my-course/team-a"}]`))
	})
	provider.cfg.ParentGroupID = &parentID

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "Team A"})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "9" {
		t.Fatalf("externalID = %q, want the existing group 9", resource.ExternalID)
	}
}

// The group search must not filter on owner access, or a maintainer token never sees
// the groups it created and tries to recreate them.
func TestGitLabGroupSearchDoesNotFilterByAccessLevel(t *testing.T) {
	rec := &requestRecorder{}
	provider := newGitLabServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":3,"path":"team-a","full_path":"team-a","web_url":"https://gitlab.test/groups/team-a"}]`))
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL)
	})

	if _, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "Team A"}); err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if strings.Contains(rec.paths[0], "min_access_level") {
		t.Fatalf("group search %q must not filter by min_access_level", rec.paths[0])
	}
}

// Members are invited by email. The old user lookup only worked for admin tokens.
func TestGitLabInvitesMembersByEmail(t *testing.T) {
	var invited url.Values
	rec := &requestRecorder{}
	provider := newGitLabServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`[{"id":3,"path":"team-a","full_path":"team-a","web_url":"https://gitlab.test/groups/team-a"}]`))
		case strings.HasSuffix(r.URL.Path, "/invitations"):
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			_ = json.Unmarshal(body, &payload)
			invited = url.Values{}
			invited.Set("email", payload["email"].(string))
			invited.Set("access_level", fmt.Sprintf("%v", payload["access_level"]))
			_, _ = w.Write([]byte(`{"status":"success"}`))
		case strings.HasSuffix(r.URL.Path, "/members"):
			t.Errorf("provider must not use the members endpoint: it needs an admin-only user lookup")
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:              "Team A",
		Members:           []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
		PermissionMapping: map[string]string{"student": "developer"},
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if len(resource.Warnings) != 0 {
		t.Fatalf("warnings = %v, want none", resource.Warnings)
	}
	if invited.Get("email") != "student@example.com" || invited.Get("access_level") != "30" {
		t.Fatalf("invitation = %v, want student@example.com at access level 30", invited)
	}
}

// A rejected invitation is reported as a warning: the group exists, so the instance is
// partial rather than failed.
func TestGitLabReportsRejectedInvitationAsWarning(t *testing.T) {
	rec := &requestRecorder{}
	provider := newGitLabServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":3,"path":"team-a","full_path":"team-a","web_url":"https://gitlab.test/groups/team-a"}]`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"error","message":{"ghost@example.com":"Invite email has already been taken"}}`))
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:    "Team A",
		Members: []providerpkg.Member{{Email: "ghost@example.com", Role: "student"}},
	})
	if err != nil {
		t.Fatalf("CreateResource must not fail when only a member could not be added: %v", err)
	}
	if len(resource.Warnings) != 1 {
		t.Fatalf("warnings = %v, want exactly one", resource.Warnings)
	}
	if resource.ExternalID != "3" {
		t.Fatalf("externalID = %q, want the created group to be reported", resource.ExternalID)
	}
}

// An already-invited member is a success, not a warning.
func TestGitLabTreatsExistingMemberAsSuccess(t *testing.T) {
	rec := &requestRecorder{}
	provider := newGitLabServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":3,"path":"team-a","full_path":"team-a","web_url":"https://gitlab.test/groups/team-a"}]`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"error","message":{"student@example.com":"Member already exists"}}`))
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:    "Team A",
		Members: []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if len(resource.Warnings) != 0 {
		t.Fatalf("warnings = %v, want none for an existing member", resource.Warnings)
	}
}

// A name that sanitizes away must be rejected before any request is sent.
func TestGitLabRejectsNameThatSanitizesToEmpty(t *testing.T) {
	rec := &requestRecorder{}
	provider := newGitLabServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected for an empty slug, got %s %s", r.Method, r.URL)
	})

	if _, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "---"}); err == nil {
		t.Fatal("CreateResource returned no error for a name that sanitizes to an empty slug")
	}
}
