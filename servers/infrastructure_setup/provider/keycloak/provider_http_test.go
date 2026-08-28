package keycloak

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	providerpkg "github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

type requestRecorder struct {
	paths []string
}

// newKeycloakServer returns a provider pointed at a test server that always issues a
// token, and delegates every admin call to handle.
func newKeycloakServer(t *testing.T, rec *requestRecorder, handle http.HandlerFunc) *Provider {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.paths = append(rec.paths, r.Method+" "+r.URL.RequestURI())
		if strings.HasSuffix(r.URL.Path, "/protocol/openid-connect/token") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"test-token","expires_in":300}`))
			return
		}
		handle(w, r)
	}))
	t.Cleanup(server.Close)

	return New(Config{KeycloakURL: server.URL, Realm: "prompt", ClientID: "c", ClientSecret: "s"})
}

func TestKeycloakCreatesGroupAndAddsMember(t *testing.T) {
	rec := &requestRecorder{}
	provider := newKeycloakServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/groups"):
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/groups"):
			w.Header().Set("Location", "https://kc.test/admin/realms/prompt/groups/group-1")
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/users"):
			_, _ = w.Write([]byte(`[{"id":"user-1"}]`))
		case r.Method == http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:         "ios2526-team-1",
		ResourceType: "group",
		Members:      []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "group-1" {
		t.Fatalf("externalID = %q, want group-1", resource.ExternalID)
	}
	if len(resource.Warnings) != 0 {
		t.Fatalf("warnings = %v, want none", resource.Warnings)
	}
	// The reported link must be openable by a human, not the bearer-protected REST path.
	if !strings.Contains(resource.ExternalURL, "/console/") {
		t.Fatalf("externalURL = %q, want an admin console URL", resource.ExternalURL)
	}
}

// An existing group of the same name is adopted, so a second run converges.
func TestKeycloakAdoptsExistingGroup(t *testing.T) {
	rec := &requestRecorder{}
	provider := newKeycloakServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.Errorf("unexpected create call: %s", r.URL)
		}
		_, _ = w.Write([]byte(`[{"id":"group-9","name":"ios2526-team-1"}]`))
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "ios2526-team-1"})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "group-9" {
		t.Fatalf("externalID = %q, want the adopted group-9", resource.ExternalID)
	}
}

// The group search must ask Keycloak to match exactly; its default search is a
// substring match whose result window can hide the group being looked for.
func TestKeycloakSearchesExactly(t *testing.T) {
	rec := &requestRecorder{}
	provider := newKeycloakServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"group-9","name":"team-1"}]`))
	})

	if _, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "team-1"}); err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	var searched string
	for _, path := range rec.paths {
		if strings.Contains(path, "/groups?") {
			searched = path
		}
	}
	if !strings.Contains(searched, "exact=true") {
		t.Fatalf("group search = %q, want exact=true", searched)
	}
}

// A create that loses a race must re-resolve the group instead of failing the instance.
func TestKeycloakRecoversFromCreateConflict(t *testing.T) {
	rec := &requestRecorder{}
	searches := 0
	provider := newKeycloakServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/groups"):
			searches++
			if searches == 1 {
				_, _ = w.Write([]byte(`[]`))
				return
			}
			_, _ = w.Write([]byte(`[{"id":"group-7","name":"team-1"}]`))
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusConflict)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "team-1"})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "group-7" {
		t.Fatalf("externalID = %q, want group-7 recovered after the conflict", resource.ExternalID)
	}
}

// A member Keycloak does not know yet is reported as a warning; the group is kept so
// the instance is partial and retryable, never silently short a member.
func TestKeycloakReportsUnknownUserAsWarning(t *testing.T) {
	rec := &requestRecorder{}
	provider := newKeycloakServer(t, rec, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/users"):
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/groups"):
			_, _ = w.Write([]byte(`[{"id":"group-1","name":"team-1"}]`))
		case r.Method == http.MethodPut:
			t.Errorf("no membership call expected for an unknown user: %s", r.URL)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:    "team-1",
		Members: []providerpkg.Member{{Email: "nobody@example.com", Role: "student"}},
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "group-1" {
		t.Fatalf("externalID = %q, want the group to be kept", resource.ExternalID)
	}
	if len(resource.Warnings) != 1 {
		t.Fatalf("warnings = %v, want one for the unknown user", resource.Warnings)
	}
}
