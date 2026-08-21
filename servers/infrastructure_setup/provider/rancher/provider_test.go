package rancher

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	providerpkg "github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

func TestRancherAuthFields(t *testing.T) {
	fields := (&Provider{}).GetAuthFields()
	if len(fields) != 4 {
		t.Fatalf("auth fields = %d, want 4", len(fields))
	}
	required := map[string]bool{}
	for _, field := range fields {
		required[field.Name] = field.Required
	}
	for _, name := range []string{"rancher_url", "access_key", "secret_key", "cluster_id"} {
		if !required[name] {
			t.Fatalf("field %q is not marked required in %+v", name, fields)
		}
	}
}

func TestNewStoresConfig(t *testing.T) {
	provider := New(Config{
		RancherURL: "https://rancher.example.com",
		AccessKey:  "access",
		SecretKey:  "secret",
		ClusterID:  "c-abcde",
	})
	if provider.cfg.ClusterID != "c-abcde" {
		t.Fatalf("ClusterID = %q, want c-abcde", provider.cfg.ClusterID)
	}
}

func newRancherServer(t *testing.T, handler http.HandlerFunc) *Provider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return New(Config{RancherURL: server.URL, AccessKey: "key", SecretKey: "secret", ClusterID: "c-abc"})
}

const rancherProjectResponse = `{"data":[{"id":"c-abc:p-1","name":"Team A","links":{"self":"https://rancher.test/p/1"}}]}`

// Rancher's user schema has no email field, so ?email= returns an unfiltered list.
// Binding the first entry would grant the local admin access to a student project.
func TestRancherDoesNotBindUnrelatedUser(t *testing.T) {
	provider := newRancherServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/v3/projects"):
			_, _ = w.Write([]byte(rancherProjectResponse))
		case r.URL.Path == "/v3/principals":
			// A fuzzy search hit that is not the requested address.
			_, _ = w.Write([]byte(`{"data":[{"id":"local://user-admin","loginName":"admin","name":"Local Admin","principalType":"user"}]}`))
		case r.URL.Path == "/v3/projectroletemplatebindings":
			t.Errorf("must not create a binding when no principal matches the email")
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:    "Team A",
		Members: []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if len(resource.Warnings) != 1 {
		t.Fatalf("warnings = %v, want one warning for the unmatched user", resource.Warnings)
	}
}

// The principal comes from principalIds, not "local://"+id, which is wrong on any
// LDAP- or OIDC-backed Rancher.
func TestRancherUsesPrincipalFromResponse(t *testing.T) {
	var boundPrincipal string
	provider := newRancherServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/v3/projects"):
			_, _ = w.Write([]byte(rancherProjectResponse))
		case r.URL.Path == "/v3/principals":
			_, _ = w.Write([]byte(`{"data":[{"id":"openldap_user://uid=student,dc=example,dc=com","loginName":"student@example.com","name":"Student","principalType":"user"}]}`))
		case r.URL.Path == "/v3/projectroletemplatebindings":
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			_ = json.Unmarshal(body, &payload)
			boundPrincipal, _ = payload["userPrincipalId"].(string)
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
	})

	if _, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:    "Team A",
		Members: []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
	}); err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if boundPrincipal != "openldap_user://uid=student,dc=example,dc=com" {
		t.Fatalf("bound principal = %q, want the principal from the user record", boundPrincipal)
	}
}

func TestRancherPrincipalMatchesEmail(t *testing.T) {
	tests := []struct {
		name      string
		loginName string
		display   string
		want      bool
	}{
		{"login name is the address", "student@example.com", "Student", true},
		{"display name is the address", "student", "student@example.com", true},
		{"unrelated principal", "admin", "Local Admin", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := principalMatchesEmail(tt.loginName, tt.display, "student@example.com"); got != tt.want {
				t.Fatalf("principalMatchesEmail = %v, want %v", got, tt.want)
			}
		})
	}
}
