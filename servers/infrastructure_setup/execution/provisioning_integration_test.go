package execution

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/encryption"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/providerconfig"
)

// These tests drive the real worker through the real provider implementations against
// mock GitLab and Outline servers, so the whole path from a resource config to an
// external call and back into the instance row is exercised. The provider unit tests
// cover request shapes; this covers the wiring between them.

// testEncryptionKey is a fixed base64-encoded 32-byte key. Credentials have to be stored
// encrypted, so a real key is needed even though nothing here is secret.
const testEncryptionKey = "aW5mcmEtc2V0dXAtaW50ZWdyYXRpb24tdGVzdC1rZXk="

func withEncryptionKey(t *testing.T) {
	t.Helper()
	previous, had := os.LookupEnv("ENCRYPTION_KEY")
	if err := os.Setenv("ENCRYPTION_KEY", testEncryptionKey); err != nil {
		t.Fatalf("set ENCRYPTION_KEY: %v", err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("ENCRYPTION_KEY", previous)
			return
		}
		_ = os.Unsetenv("ENCRYPTION_KEY")
	})
}

// callLog records every request a mock upstream received.
type callLog struct {
	mu    sync.Mutex
	calls []string
}

func (c *callLog) record(entry string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, entry)
}

func (c *callLog) all() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.calls...)
}

func (c *callLog) contains(substr string) bool {
	for _, entry := range c.all() {
		if strings.Contains(entry, substr) {
			return true
		}
	}
	return false
}

// registerRealProvider points the execution registry at the genuine provider factory,
// restoring whatever was there afterwards.
func registerRealProvider(t *testing.T, providerType string) {
	t.Helper()
	previous, had := Registry[providerType]
	Registry[providerType] = func(creds []byte) (provider.Provider, error) {
		return providerconfig.BuildProviderFromEncryptedCreds(providerType, creds)
	}
	t.Cleanup(func() {
		if had {
			Registry[providerType] = previous
			return
		}
		delete(Registry, providerType)
	})
}

func storeCredentials(t *testing.T, queries *db.Queries, coursePhaseID uuid.UUID, providerType db.ProviderType, credentials map[string]string) {
	t.Helper()
	raw, err := json.Marshal(credentials)
	if err != nil {
		t.Fatalf("marshal credentials: %v", err)
	}
	encrypted, err := encryption.Encrypt(raw)
	if err != nil {
		t.Fatalf("encrypt credentials: %v", err)
	}
	if _, err := queries.UpsertProviderConfig(context.Background(), db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  providerType,
		Credentials:   encrypted,
	}); err != nil {
		t.Fatalf("upsert provider config: %v", err)
	}
}

func teamTarget(name string, emails ...string) ProvisioningTarget {
	teamID := uuid.New()
	members := make([]provider.Member, 0, len(emails))
	for _, email := range emails {
		members = append(members, provider.Member{Email: email, Role: "student"})
	}
	return ProvisioningTarget{
		Scope:        db.ResourceScopePerTeam,
		TeamID:       &teamID,
		TeamName:     name,
		Members:      members,
		TemplateData: TemplateData{TeamName: name, SemesterTag: "ios2526"},
	}
}

// runWorker executes the pending instances synchronously so the test can assert on the
// end state without polling a goroutine.
func runWorker(t *testing.T, worker *Worker, coursePhaseID uuid.UUID) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := worker.processPhase(ctx, "Bearer test", coursePhaseID); err != nil {
		t.Fatalf("processPhase: %v", err)
	}
}

func instancesByStatus(t *testing.T, queries *db.Queries, coursePhaseID uuid.UUID) map[db.ResourceStatus][]db.ResourceInstance {
	t.Helper()
	instances, err := queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	byStatus := map[db.ResourceStatus][]db.ResourceInstance{}
	for _, inst := range instances {
		byStatus[inst.Status] = append(byStatus[inst.Status], inst)
	}
	return byStatus
}

// A GitLab project config provisions the team subgroup and the repository inside it, and
// records the project on the instance.
func TestProvisioningCreatesGitLabProjectForTeam(t *testing.T) {
	withEncryptionKey(t)
	registerRealProvider(t, "gitlab")

	log := &callLog{}
	var projectPayload map[string]interface{}
	gitlab := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(r.Method + " " + r.RequestURI)
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/subgroups"):
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v4/groups":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":500,"web_url":"https://gitlab.test/groups/ios2526/ios2526-team-1"}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/invitations"):
			_, _ = w.Write([]byte(`{"status":"success"}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v4/projects/"):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"404 Project Not Found"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v4/projects":
			_ = json.NewDecoder(r.Body).Decode(&projectPayload)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":900,"web_url":"https://gitlab.test/ios2526/ios2526-team-1/app"}`))
		default:
			t.Errorf("unexpected GitLab request: %s %s", r.Method, r.RequestURI)
		}
	}))
	defer gitlab.Close()

	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()
	queries := testDB.Queries
	coursePhaseID := uuid.New()

	storeCredentials(t, queries, coursePhaseID, db.ProviderTypeGitlab, map[string]string{
		"base_url":        gitlab.URL,
		"private_token":   "test-token",
		"parent_group_id": "42",
	})

	cfg, err := queries.CreateResourceConfig(context.Background(), db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        db.ProviderTypeGitlab,
		ResourceType:        "project",
		Scope:               db.ResourceScopePerTeam,
		NameTemplate:        "{{semesterTag}}-{{teamName}}-app",
		PermissionMapping:   []byte(`{"student":"developer"}`),
		ResourceExtraConfig: []byte(`{"parent_group_template":"{{semesterTag}}-{{teamName}}"}`),
	})
	if err != nil {
		t.Fatalf("create resource config: %v", err)
	}

	targets := []ProvisioningTarget{teamTarget("Team 1", "student@example.com")}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})
	if err := createInstances(t, service, coursePhaseID, cfg, targets); err != nil {
		t.Fatalf("create instances: %v", err)
	}

	runWorker(t, NewWorkerWithResolver(testDB.Conn, fakeTargetResolver{targets: targets}), coursePhaseID)

	byStatus := instancesByStatus(t, queries, coursePhaseID)
	created := byStatus[db.ResourceStatusCreated]
	if len(created) != 1 {
		t.Fatalf("instances by status = %v, want exactly one created", summarize(byStatus))
	}
	if created[0].ExternalID == nil || *created[0].ExternalID != "900" {
		t.Fatalf("externalID = %v, want the project 900", created[0].ExternalID)
	}

	// The extra-config template must have been resolved per team before the call.
	if projectPayload["namespace_id"] != float64(500) {
		t.Fatalf("namespace_id = %v, want the created team subgroup 500", projectPayload["namespace_id"])
	}
	if projectPayload["path"] != "ios2526-team-1-app" {
		t.Fatalf("project path = %v, want the resolved name template", projectPayload["path"])
	}
	if !log.contains("POST /api/v4/groups/500/invitations") {
		t.Fatalf("members were not invited to the subgroup; calls: %v", log.all())
	}
}

// An Outline collection config provisions a private collection plus the group bound to it.
func TestProvisioningBindsOutlineGroupToCollection(t *testing.T) {
	withEncryptionKey(t)
	registerRealProvider(t, "outline")

	log := &callLog{}
	payloads := map[string]map[string]interface{}{}
	var mu sync.Mutex
	outline := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := strings.TrimPrefix(r.URL.Path, "/")
		log.record("POST " + method)

		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		mu.Lock()
		payloads[method] = payload
		mu.Unlock()

		switch method {
		case "collections.list":
			_, _ = w.Write([]byte(`{"ok":true,"data":[]}`))
		case "collections.create":
			_, _ = w.Write([]byte(`{"ok":true,"data":{"collection":{"id":"col-1","url":"/collection/team-1"}}}`))
		case "groups.list":
			_, _ = w.Write([]byte(`{"ok":true,"data":{"groups":[]}}`))
		case "groups.create":
			_, _ = w.Write([]byte(`{"ok":true,"data":{"id":"grp-1"}}`))
		case "users.list":
			_, _ = w.Write([]byte(`{"ok":true,"data":[{"id":"usr-1","email":"student@example.com"}]}`))
		case "groups.add_user", "collections.add_group":
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			t.Errorf("unexpected Outline call: %s", method)
			_, _ = w.Write([]byte(`{"ok":false}`))
		}
	}))
	defer outline.Close()

	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()
	queries := testDB.Queries
	coursePhaseID := uuid.New()

	storeCredentials(t, queries, coursePhaseID, db.ProviderTypeOutline, map[string]string{
		"api_key":  "ol_api_test",
		"base_url": outline.URL,
	})

	cfg, err := queries.CreateResourceConfig(context.Background(), db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        db.ProviderTypeOutline,
		ResourceType:        "collection",
		Scope:               db.ResourceScopePerTeam,
		NameTemplate:        "{{semesterTag}}-{{teamName}}",
		PermissionMapping:   []byte(`{"student":"read"}`),
		ResourceExtraConfig: []byte(`{"group_name_template":"{{semesterTag}}-{{teamName}}"}`),
	})
	if err != nil {
		t.Fatalf("create resource config: %v", err)
	}

	targets := []ProvisioningTarget{teamTarget("Team 1", "student@example.com")}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})
	if err := createInstances(t, service, coursePhaseID, cfg, targets); err != nil {
		t.Fatalf("create instances: %v", err)
	}

	runWorker(t, NewWorkerWithResolver(testDB.Conn, fakeTargetResolver{targets: targets}), coursePhaseID)

	byStatus := instancesByStatus(t, queries, coursePhaseID)
	created := byStatus[db.ResourceStatusCreated]
	if len(created) != 1 {
		t.Fatalf("instances by status = %v, want exactly one created", summarize(byStatus))
	}
	if created[0].ExternalID == nil || *created[0].ExternalID != "col-1" {
		t.Fatalf("externalID = %v, want the collection col-1", created[0].ExternalID)
	}

	mu.Lock()
	defer mu.Unlock()
	if _, present := payloads["collections.create"]["permission"]; present {
		t.Fatal("the collection was created with a permission, which makes it workspace-readable")
	}
	// The group must be keyed on the instance's stable key, which embeds the phase and
	// config IDs rather than the display name.
	externalID, _ := payloads["groups.create"]["externalId"].(string)
	if !strings.HasPrefix(externalID, "prompt:"+coursePhaseID.String()+":"+cfg.ID.String()+":") {
		t.Fatalf("groups.create externalId = %q, want the instance's stable key", externalID)
	}
	if payloads["collections.add_group"]["permission"] != "read" {
		t.Fatalf("collections.add_group permission = %v, want an explicit read",
			payloads["collections.add_group"]["permission"])
	}
}

// A member the upstream rejects leaves the instance partial with its external ID intact,
// and retrying heals it once the member resolves. This is the path a fresh cohort hits,
// where students have not signed in to the upstream system yet.
func TestProvisioningRecordsPartialAndHealsOnRetry(t *testing.T) {
	withEncryptionKey(t)
	registerRealProvider(t, "gitlab")

	var rejectInvites atomic.Bool
	rejectInvites.Store(true)

	gitlab := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/groups"):
			_, _ = w.Write([]byte(`[{"id":600,"path":"ios2526-team-1","full_path":"ios2526-team-1","web_url":"https://gitlab.test/groups/ios2526-team-1"}]`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/invitations"):
			if rejectInvites.Load() {
				_, _ = w.Write([]byte(`{"status":"error","message":{"student@example.com":"Invite email is invalid"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"status":"success"}`))
		default:
			t.Errorf("unexpected GitLab request: %s %s", r.Method, r.RequestURI)
		}
	}))
	defer gitlab.Close()

	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()
	queries := testDB.Queries
	coursePhaseID := uuid.New()

	storeCredentials(t, queries, coursePhaseID, db.ProviderTypeGitlab, map[string]string{
		"base_url":      gitlab.URL,
		"private_token": "test-token",
	})

	cfg, err := queries.CreateResourceConfig(context.Background(), db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        db.ProviderTypeGitlab,
		ResourceType:        "group",
		Scope:               db.ResourceScopePerTeam,
		NameTemplate:        "{{semesterTag}}-{{teamName}}",
		PermissionMapping:   []byte(`{"student":"developer"}`),
		ResourceExtraConfig: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("create resource config: %v", err)
	}

	targets := []ProvisioningTarget{teamTarget("Team 1", "student@example.com")}
	resolver := fakeTargetResolver{targets: targets}
	service := NewServiceWithResolver(testDB.Conn, resolver)
	if err := createInstances(t, service, coursePhaseID, cfg, targets); err != nil {
		t.Fatalf("create instances: %v", err)
	}

	worker := NewWorkerWithResolver(testDB.Conn, resolver)
	runWorker(t, worker, coursePhaseID)

	partial := instancesByStatus(t, queries, coursePhaseID)[db.ResourceStatusPartial]
	if len(partial) != 1 {
		t.Fatalf("instances = %v, want one partial", summarize(instancesByStatus(t, queries, coursePhaseID)))
	}
	// The group exists, so its ID must survive the member failure.
	if partial[0].ExternalID == nil || *partial[0].ExternalID != "600" {
		t.Fatalf("externalID = %v, want the group to be retained on a partial instance", partial[0].ExternalID)
	}
	if partial[0].ErrorMessage == nil || !strings.Contains(*partial[0].ErrorMessage, "student@example.com") {
		t.Fatalf("errorMessage = %v, want it to name the member that failed", partial[0].ErrorMessage)
	}

	// Retry once the upstream accepts the member.
	rejectInvites.Store(false)
	if err := service.RetryInstance(context.Background(), "Bearer test", coursePhaseID, partial[0].ID); err != nil {
		t.Fatalf("retry instance: %v", err)
	}
	runWorker(t, worker, coursePhaseID)

	byStatus := instancesByStatus(t, queries, coursePhaseID)
	if len(byStatus[db.ResourceStatusCreated]) != 1 || len(byStatus[db.ResourceStatusPartial]) != 0 {
		t.Fatalf("after retry instances = %v, want one created and no partial", summarize(byStatus))
	}
}

// summarize renders the instance statuses for a readable failure message.
func summarize(byStatus map[db.ResourceStatus][]db.ResourceInstance) map[string]int {
	out := map[string]int{}
	for status, instances := range byStatus {
		out[string(status)] = len(instances)
	}
	return out
}
