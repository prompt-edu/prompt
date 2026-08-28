package outline

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
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

// recorder captures the API methods called, and the payload of each, so a test can
// assert both the sequence and the exact request bodies.
type recorder struct {
	methods  []string
	payloads map[string]map[string]interface{}
}

func newRecorder() *recorder {
	return &recorder{payloads: map[string]map[string]interface{}{}}
}

// newRecordingProvider serves a scripted response per Outline method. A method with no
// entry in responses fails the test, so an unexpected call cannot pass silently.
func newRecordingProvider(t *testing.T, rec *recorder, responses map[string]string) *Provider {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := strings.TrimPrefix(r.URL.Path, "/")
		rec.methods = append(rec.methods, method)

		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		rec.payloads[method] = payload

		body, ok := responses[method]
		if !ok {
			t.Errorf("unexpected Outline call: %s", method)
			body = `{"ok":false,"error":"unexpected"}`
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return New(Config{APIKey: "ol_api_test", BaseURL: server.URL})
}

const stableKey = "prompt:phase-1:config-1:team-1"

// The whole happy path: a private collection, a group keyed on the stable external ID,
// the members in that group, and the group bound to the collection.
func TestOutlineCreatesPrivateCollectionAndBindsGroup(t *testing.T) {
	rec := newRecorder()
	provider := newRecordingProvider(t, rec, map[string]string{
		"collections.list":      `{"ok":true,"data":[]}`,
		"collections.create":    `{"ok":true,"data":{"collection":{"id":"col-1","url":"/collection/team-a"}}}`,
		"groups.list":           `{"ok":true,"data":{"groups":[]}}`,
		"groups.create":         `{"ok":true,"data":{"id":"grp-1"}}`,
		"users.list":            `{"ok":true,"data":[{"id":"usr-1","email":"student@example.com"}]}`,
		"groups.add_user":       `{"ok":true}`,
		"collections.add_group": `{"ok":true}`,
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:              "Team A",
		ResourceType:      "collection",
		Members:           []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
		PermissionMapping: map[string]string{"student": "read"},
		StableKey:         stableKey,
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "col-1" {
		t.Fatalf("externalID = %q, want the collection col-1", resource.ExternalID)
	}
	if len(resource.Warnings) != 0 {
		t.Fatalf("warnings = %v, want none", resource.Warnings)
	}

	// permission must be absent, otherwise every workspace member can read the collection.
	create := rec.payloads["collections.create"]
	if _, present := create["permission"]; present {
		t.Fatalf("collections.create sent permission=%v; a private collection must omit it", create["permission"])
	}
	if !strings.Contains(fmt.Sprint(create["description"]), stableKey) {
		t.Fatalf("description = %v, want the ownership marker", create["description"])
	}

	// The group is keyed on the stable ID, not on the display name.
	if rec.payloads["groups.create"]["externalId"] != stableKey+":read" {
		t.Fatalf("groups.create externalId = %v, want %q", rec.payloads["groups.create"]["externalId"], stableKey+":read")
	}

	// The binding permission must be explicit: Outline defaults add_group to read_write.
	bind := rec.payloads["collections.add_group"]
	if bind["permission"] != "read" {
		t.Fatalf("collections.add_group permission = %v, want an explicit read", bind["permission"])
	}
	if bind["groupId"] != "grp-1" || bind["id"] != "col-1" {
		t.Fatalf("collections.add_group = %v, want group grp-1 on collection col-1", bind)
	}
}

// A same-named collection this phase did not create must not be adopted: Outline
// collections share one flat namespace, so adopting one would hand the team read access
// to unrelated documents.
func TestOutlineRefusesUnrelatedCollectionOfSameName(t *testing.T) {
	rec := newRecorder()
	provider := newRecordingProvider(t, rec, map[string]string{
		"collections.list": `{"ok":true,"data":[{"id":"col-9","name":"Team A","url":"/c/x","description":"Someone else's notes"}]}`,
	})

	_, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:      "Team A",
		StableKey: stableKey,
	})
	if err == nil {
		t.Fatal("CreateResource = nil error, want a refusal to adopt an unrelated collection")
	}
	if slices.Contains(rec.methods, "collections.create") {
		t.Fatal("a duplicate collection was created instead of refusing")
	}
}

// A collection this phase created is adopted, so a re-run converges.
func TestOutlineAdoptsOwnCollection(t *testing.T) {
	rec := newRecorder()
	provider := newRecordingProvider(t, rec, map[string]string{
		"collections.list": fmt.Sprintf(
			`{"ok":true,"data":[{"id":"col-7","name":"Team A","url":"/c/team-a","description":"Provisioned by PROMPT.\n\nprompt-key: %s"}]}`, stableKey),
		"groups.list":           `{"ok":true,"data":{"groups":[{"id":"grp-7","externalId":"` + stableKey + `:read"}]}}`,
		"users.list":            `{"ok":true,"data":[]}`,
		"groups.add_user":       `{"ok":true}`,
		"collections.add_group": `{"ok":true}`,
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:              "Team A",
		Members:           []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
		PermissionMapping: map[string]string{"student": "read"},
		StableKey:         stableKey,
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "col-7" {
		t.Fatalf("externalID = %q, want the adopted collection col-7", resource.ExternalID)
	}
	if slices.Contains(rec.methods, "collections.create") {
		t.Fatal("an owned collection was recreated instead of adopted")
	}
	// The existing group is reused rather than duplicated.
	if slices.Contains(rec.methods, "groups.create") {
		t.Fatal("an existing group was recreated instead of adopted")
	}
}

// A retried instance re-attaches by the recorded ID, so a renamed collection is still
// found and the name never decides anything.
func TestOutlineReattachesByExistingExternalID(t *testing.T) {
	rec := newRecorder()
	provider := newRecordingProvider(t, rec, map[string]string{
		"collections.list":      `{"ok":true,"data":[{"id":"col-5","name":"Renamed By A Lecturer","url":"/c/renamed"}]}`,
		"groups.list":           `{"ok":true,"data":{"groups":[]}}`,
		"groups.create":         `{"ok":true,"data":{"id":"grp-5"}}`,
		"users.list":            `{"ok":true,"data":[]}`,
		"collections.add_group": `{"ok":true}`,
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:               "Team A",
		Members:            []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
		PermissionMapping:  map[string]string{"student": "read"},
		StableKey:          stableKey,
		ExistingExternalID: "col-5",
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "col-5" {
		t.Fatalf("externalID = %q, want the recorded col-5", resource.ExternalID)
	}
	if slices.Contains(rec.methods, "collections.create") {
		t.Fatal("the recorded collection was recreated instead of re-attached")
	}
}

// Members whose roles map to different permissions need one group each: a group carries
// a single collection permission, so choosing one silently would over- or under-grant.
func TestOutlineSplitsGroupsPerPermission(t *testing.T) {
	rec := newRecorder()
	var groupNames []string
	var boundPermissions []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := strings.TrimPrefix(r.URL.Path, "/")
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		rec.methods = append(rec.methods, method)

		switch method {
		case "collections.list":
			_, _ = w.Write([]byte(`{"ok":true,"data":[]}`))
		case "collections.create":
			_, _ = w.Write([]byte(`{"ok":true,"data":{"collection":{"id":"col-1","url":"/c/1"}}}`))
		case "users.list":
			_, _ = w.Write([]byte(`{"ok":true,"data":[{"id":"usr-1","email":"student@example.com"},{"id":"usr-2","email":"tutor@example.com"}]}`))
		case "groups.list":
			_, _ = w.Write([]byte(`{"ok":true,"data":{"groups":[]}}`))
		case "groups.create":
			groupNames = append(groupNames, fmt.Sprint(payload["name"]))
			_, _ = w.Write([]byte(fmt.Sprintf(`{"ok":true,"data":{"id":"grp-%d"}}`, len(groupNames))))
		case "groups.add_user":
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "collections.add_group":
			boundPermissions = append(boundPermissions, fmt.Sprint(payload["permission"]))
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			t.Errorf("unexpected Outline call: %s", method)
		}
	}))
	t.Cleanup(server.Close)
	provider := New(Config{APIKey: "k", BaseURL: server.URL})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name: "Team A",
		Members: []providerpkg.Member{
			{Email: "student@example.com", Role: "student"},
			{Email: "tutor@example.com", Role: "tutor"},
		},
		PermissionMapping: map[string]string{"student": "read", "tutor": "read_write"},
		StableKey:         stableKey,
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if len(resource.Warnings) != 0 {
		t.Fatalf("warnings = %v, want none", resource.Warnings)
	}
	if len(groupNames) != 2 {
		t.Fatalf("groups created = %v, want one per permission", groupNames)
	}
	// Split groups need distinguishable display names, or Outline's group list is unreadable.
	for _, name := range groupNames {
		if !strings.Contains(name, "read") {
			t.Fatalf("group name %q does not name its permission", name)
		}
	}
	if len(groupNames) != len(map[string]bool{groupNames[0]: true, groupNames[1]: true}) {
		t.Fatalf("group names are not distinct: %v", groupNames)
	}
	slices.Sort(boundPermissions)
	if !slices.Equal(boundPermissions, []string{"read", "read_write"}) {
		t.Fatalf("bound permissions = %v, want read and read_write", boundPermissions)
	}
}

// Once the collection exists, a failing bind is a warning: the instance must keep its
// external ID so Retry re-attaches by ID rather than re-adopting by name.
func TestOutlineKeepsCollectionWhenBindFails(t *testing.T) {
	rec := newRecorder()
	provider := newRecordingProvider(t, rec, map[string]string{
		"collections.list":      `{"ok":true,"data":[]}`,
		"collections.create":    `{"ok":true,"data":{"collection":{"id":"col-1","url":"/c/1"}}}`,
		"groups.list":           `{"ok":true,"data":{"groups":[]}}`,
		"groups.create":         `{"ok":true,"data":{"id":"grp-1"}}`,
		"users.list":            `{"ok":true,"data":[{"id":"usr-1","email":"student@example.com"}]}`,
		"groups.add_user":       `{"ok":true}`,
		"collections.add_group": `{"ok":false,"error":"forbidden"}`,
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:              "Team A",
		Members:           []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
		PermissionMapping: map[string]string{"student": "read"},
		StableKey:         stableKey,
	})
	if err != nil {
		t.Fatalf("CreateResource returned an error; a created collection must be kept: %v", err)
	}
	if resource.ExternalID != "col-1" {
		t.Fatalf("externalID = %q, want the created collection to be retained", resource.ExternalID)
	}
	if len(resource.Warnings) != 1 {
		t.Fatalf("warnings = %v, want one for the failed binding", resource.Warnings)
	}
}

// A role with no mapping must be reported, not quietly granted read access.
func TestOutlineReportsUnmappedRole(t *testing.T) {
	rec := newRecorder()
	provider := newRecordingProvider(t, rec, map[string]string{
		"collections.list":   `{"ok":true,"data":[]}`,
		"collections.create": `{"ok":true,"data":{"collection":{"id":"col-1","url":"/c/1"}}}`,
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:      "Team A",
		Members:   []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
		StableKey: stableKey,
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if len(resource.Warnings) != 1 {
		t.Fatalf("warnings = %v, want one for the unmapped role", resource.Warnings)
	}
	if slices.Contains(rec.methods, "groups.create") {
		t.Fatal("a group was created for a member with no mapped permission")
	}
}

// An existing collection on a later page must be found, or a duplicate is created.
func TestOutlineFindsCollectionOnLaterPage(t *testing.T) {
	rec := newRecorder()
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := strings.TrimPrefix(r.URL.Path, "/")
		rec.methods = append(rec.methods, method)
		switch method {
		case "collections.list":
			page++
			if page == 1 {
				entries := make([]string, 0, outlinePageSize)
				for i := 0; i < outlinePageSize; i++ {
					entries = append(entries, fmt.Sprintf(`{"id":"other-%d","name":"Other %d","url":"/c/o"}`, i, i))
				}
				_, _ = w.Write([]byte(`{"ok":true,"data":[` + strings.Join(entries, ",") + `]}`))
				return
			}
			_, _ = w.Write([]byte(fmt.Sprintf(
				`{"ok":true,"data":[{"id":"col-42","name":"Team A","url":"/c/team-a","description":"prompt-key: %s"}]}`, stableKey)))
		case "groups.list":
			_, _ = w.Write([]byte(`{"ok":true,"data":{"groups":[{"id":"grp-42","externalId":"` + stableKey + `:read"}]}}`))
		case "users.list":
			_, _ = w.Write([]byte(`{"ok":true,"data":[]}`))
		case "collections.add_group":
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "collections.create":
			t.Error("a duplicate collection was created despite the existing one on page 2")
		default:
			t.Errorf("unexpected Outline call: %s", method)
		}
	}))
	t.Cleanup(server.Close)
	provider := New(Config{APIKey: "k", BaseURL: server.URL})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:              "Team A",
		Members:           []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
		PermissionMapping: map[string]string{"student": "read"},
		StableKey:         stableKey,
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if resource.ExternalID != "col-42" {
		t.Fatalf("externalID = %q, want the collection from page 2", resource.ExternalID)
	}
}

func TestOutlineRejectsEmptyName(t *testing.T) {
	provider := newOutlineProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected for an empty name: %s", r.URL.Path)
	})

	if _, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{Name: "   "}); err == nil {
		t.Fatal("CreateResource = nil error, want a rejection for an empty name")
	}
}

// The group key must not depend on how many permissions a team happens to have. If the
// lone group were keyed on the bare stable key, adding a second permission later would
// change the key and orphan the existing group.
func TestOutlineGroupKeyAlwaysCarriesThePermission(t *testing.T) {
	rec := newRecorder()
	provider := newRecordingProvider(t, rec, map[string]string{
		"collections.list":      `{"ok":true,"data":[]}`,
		"collections.create":    `{"ok":true,"data":{"collection":{"id":"col-1","url":"/c/1"}}}`,
		"groups.list":           `{"ok":true,"data":{"groups":[{"id":"grp-1","externalId":"` + stableKey + `:read"}]}}`,
		"users.list":            `{"ok":true,"data":[{"id":"usr-1","email":"student@example.com"}]}`,
		"groups.add_user":       `{"ok":true}`,
		"collections.add_group": `{"ok":true}`,
	})

	resource, err := provider.CreateResource(context.Background(), providerpkg.CreateResourceInput{
		Name:              "Team A",
		Members:           []providerpkg.Member{{Email: "student@example.com", Role: "student"}},
		PermissionMapping: map[string]string{"student": "read"},
		StableKey:         stableKey,
	})
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if len(resource.Warnings) != 0 {
		t.Fatalf("warnings = %v, want none", resource.Warnings)
	}
	// A single-permission team must still look the group up under the suffixed key, and
	// therefore adopt the existing one rather than creating a second.
	if rec.payloads["groups.list"]["externalId"] != stableKey+":read" {
		t.Fatalf("groups.list externalId = %v, want the permission-suffixed key",
			rec.payloads["groups.list"]["externalId"])
	}
	if slices.Contains(rec.methods, "groups.create") {
		t.Fatal("the existing group was not adopted under the suffixed key")
	}
}
