package outline

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	providerpkg "github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

func TestNewDefaultsBaseURL(t *testing.T) {
	provider := New(Config{APIKey: "secret"})
	if provider.cfg.BaseURL != "https://app.getoutline.com/api" {
		t.Fatalf("BaseURL = %q, want Outline cloud API default", provider.cfg.BaseURL)
	}
}

func TestOutlineAuthFields(t *testing.T) {
	fields := (&Provider{}).GetAuthFields()
	if len(fields) != 2 {
		t.Fatalf("auth fields = %d, want 2", len(fields))
	}
	if fields[0].Name != "api_key" || fields[0].Type != "password" || !fields[0].Required {
		t.Fatalf("first auth field = %+v, want required api_key password", fields[0])
	}
}

func newOutlineProvider(t *testing.T, handler http.HandlerFunc) *Provider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return New(Config{APIKey: "ol_api_test", BaseURL: server.URL})
}

// An ok:false on the membership call must surface. It was ignored, so a resource was
// reported as fully provisioned while nobody had been added.
func TestOutlineReportsFailedMembershipAsWarning(t *testing.T) {
	provider := newOutlineProvider(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "collections.list"):
			_, _ = w.Write([]byte(`{"ok":true,"data":[{"id":"col-1","name":"Team A","url":"/collection/team-a"}]}`))
		case strings.HasSuffix(r.URL.Path, "users.list"):
			_, _ = w.Write([]byte(`{"ok":true,"data":[{"id":"u-1","email":"student@example.com"}]}`))
		case strings.HasSuffix(r.URL.Path, "collections.add_user"):
			_, _ = w.Write([]byte(`{"ok":false,"error":"permission_denied"}`))
		default:
			t.Errorf("unexpected call: %s", r.URL.Path)
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
		t.Fatalf("warnings = %v, want one for the rejected membership", resource.Warnings)
	}
}

// A collection beyond the first page must be found, or it gets recreated under a
// duplicate name.
func TestOutlineFindsCollectionOnLaterPage(t *testing.T) {
	listCalls := 0
	provider := newOutlineProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "collections.create") {
			t.Errorf("must not create a collection that already exists on a later page")
		}
		if !strings.HasSuffix(r.URL.Path, "collections.list") {
			t.Errorf("unexpected call: %s", r.URL.Path)
			return
		}
		listCalls++
		if listCalls == 1 {
			// A full page of unrelated collections.
			entries := make([]string, 0, outlinePageSize)
			for i := range outlinePageSize {
				entries = append(entries, fmt.Sprintf(`{"id":"other-%d","name":"Other %d","url":"/o"}`, i, i))
			}
			_, _ = fmt.Fprintf(w, `{"ok":true,"data":[%s]}`, strings.Join(entries, ","))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"data":[{"id":"col-42","name":"Team A","url":"/collection/team-a"}]}`))
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "Team A"})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "col-42" {
		t.Fatalf("externalID = %q, want col-42 from the second page", resource.ExternalID)
	}
}

func TestOutlineRejectsEmptyName(t *testing.T) {
	provider := newOutlineProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected for an empty name, got %s", r.URL.Path)
	})
	if _, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "  "}); err == nil {
		t.Fatal("CreateResource returned no error for an empty name")
	}
}
