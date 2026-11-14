package keycloakRealmManager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Nerzal/gocloak/v14"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeKeycloak serves the admin endpoints the org functions call and records the writes.
type fakeKeycloak struct {
	mu              sync.Mutex
	groupIDs        map[string]string   // group path → group ID
	members         map[string][]string // group ID → user IDs
	users           []gocloak.User      // the realm's users, returned with their attributes
	loginFails      bool
	membersFail     bool
	usersFailingAdd map[string]bool
	unknownUsers    map[string]bool // user IDs that Keycloak answers with 404
	deleted         []string
	added           []string
	writeLog        []string // every write in order, as "PUT <path>" or "DELETE <path>"
}

func (f *fakeKeycloak) user(userID string) gocloak.User {
	for _, user := range f.users {
		if *user.ID == userID {
			return user
		}
	}
	return gocloak.User{ID: gocloak.StringP(userID)}
}

func (f *fakeKeycloak) start(t *testing.T) *KeycloakRealmService {
	writeJSON := func(w http.ResponseWriter, body any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /realms/prompt/protocol/openid-connect/token", func(w http.ResponseWriter, _ *http.Request) {
		if f.loginFails {
			http.Error(w, `{"error":"invalid_client"}`, http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]any{"access_token": "token", "expires_in": 60})
	})
	mux.HandleFunc("GET /admin/realms/prompt/group-by-path/{path...}", func(w http.ResponseWriter, r *http.Request) {
		groupPath := r.PathValue("path")
		groupID, ok := f.groupIDs[groupPath]
		if !ok {
			http.Error(w, `{"error":"Group path does not exist"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, gocloak.Group{ID: &groupID, Name: gocloak.StringP(path.Base(groupPath))})
	})
	mux.HandleFunc("GET /admin/realms/prompt/groups/{id}/members", func(w http.ResponseWriter, r *http.Request) {
		if f.membersFail {
			http.Error(w, `{"error":"unavailable"}`, http.StatusInternalServerError)
			return
		}
		all := f.members[r.PathValue("id")]
		first, _ := strconv.Atoi(r.URL.Query().Get("first"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("max"))
		page := []gocloak.User{}
		for _, userID := range all[min(first, len(all)):min(first+pageSize, len(all))] {
			page = append(page, f.user(userID))
		}
		writeJSON(w, page)
	})
	// Like Keycloak, the search matches substrings unless exact=true is passed.
	mux.HandleFunc("GET /admin/realms/prompt/users", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		found := []gocloak.User{}
		for _, user := range f.users {
			if attribute, value, ok := strings.Cut(query.Get("q"), ":"); ok {
				if slices.ContainsFunc(user.Attributes[attribute], func(v string) bool { return strings.Contains(v, value) }) {
					found = append(found, user)
				}
				continue
			}
			username := query.Get("username")
			if username != "" && (*user.Username == username || (query.Get("exact") != "true" && strings.Contains(*user.Username, username))) {
				found = append(found, user)
			}
		}
		writeJSON(w, found)
	})
	mux.HandleFunc("DELETE /admin/realms/prompt/", func(w http.ResponseWriter, r *http.Request) {
		resource := strings.TrimPrefix(r.URL.Path, "/admin/realms/prompt/")
		f.mu.Lock()
		f.deleted = append(f.deleted, resource)
		f.writeLog = append(f.writeLog, "DELETE "+resource)
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("DELETE /admin/realms/prompt/users/{user}/groups/{group}", func(w http.ResponseWriter, r *http.Request) {
		resource := strings.TrimPrefix(r.URL.Path, "/admin/realms/prompt/")
		f.mu.Lock()
		f.deleted = append(f.deleted, resource)
		f.writeLog = append(f.writeLog, "DELETE "+resource)
		f.mu.Unlock()
		if f.unknownUsers[r.PathValue("user")] {
			http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("PUT /admin/realms/prompt/users/{user}/groups/{group}", func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("user")
		f.mu.Lock()
		f.added = append(f.added, userID+" -> "+r.PathValue("group"))
		f.writeLog = append(f.writeLog, "PUT "+strings.TrimPrefix(r.URL.Path, "/admin/realms/prompt/"))
		f.mu.Unlock()
		if f.usersFailingAdd[userID] || f.unknownUsers[userID] {
			http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return NewKeycloakRealmService(server.URL, "prompt", "prompt-server", "secret", "client-uuid", db.Queries{})
}

func (f *fakeKeycloak) writes() (deleted, added []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.deleted, f.added
}

func orgGroupIDs() map[string]string {
	return map[string]string{
		"Orgs/lab":        "org-group",
		"Orgs/lab/Admin":  "admin-group",
		"Orgs/lab/Member": "member-group",
	}
}

func TestDeleteOrgGroupsAndRolesReturnsEveryMemberAcrossPages(t *testing.T) {
	admins := make([]string, groupMembersPageSize+50)
	for i := range admins {
		admins[i] = fmt.Sprintf("admin-%03d", i)
	}
	keycloak := &fakeKeycloak{
		groupIDs: orgGroupIDs(),
		members:  map[string][]string{"admin-group": admins, "member-group": {"member-1"}},
	}

	members, err := keycloak.start(t).DeleteOrgGroupsAndRoles(context.Background(), "lab")

	require.NoError(t, err)
	assert.Equal(t, admins, members[permissionValidation.OrgAdmin])
	assert.Equal(t, []string{"member-1"}, members[permissionValidation.OrgMember])
	deleted, _ := keycloak.writes()
	assert.Equal(t, []string{
		"clients/client-uuid/roles/org-lab-Admin",
		"clients/client-uuid/roles/org-lab-Member",
		"groups/org-group",
	}, deleted)
}

func TestDeleteOrgGroupsAndRolesChangesNothingWhenTheMembersCannotBeRead(t *testing.T) {
	keycloak := &fakeKeycloak{groupIDs: orgGroupIDs(), membersFail: true}

	members, err := keycloak.start(t).DeleteOrgGroupsAndRoles(context.Background(), "lab")

	assert.ErrorContains(t, err, "failed to get the members of the Admin group")
	assert.Nil(t, members)
	deleted, _ := keycloak.writes()
	assert.Empty(t, deleted)
}

func TestAddOrgMembersReAddsEveryUserAndReportsTheOnesItCouldNot(t *testing.T) {
	keycloak := &fakeKeycloak{groupIDs: orgGroupIDs(), usersFailingAdd: map[string]bool{"member-2": true}}

	err := keycloak.start(t).AddOrgMembers(context.Background(), "lab", map[string][]string{
		permissionValidation.OrgAdmin:  {"admin-1"},
		permissionValidation.OrgMember: {"member-1", "member-2"},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "member-2")
	assert.NotContains(t, err.Error(), "admin-1")
	assert.NotContains(t, err.Error(), "member-1")
	_, added := keycloak.writes()
	assert.Equal(t, []string{"admin-1 -> admin-group", "member-1 -> member-group", "member-2 -> member-group"}, added)
}

func TestAddOrgMembersNamesEveryMemberWhenTheLoginFails(t *testing.T) {
	keycloak := &fakeKeycloak{groupIDs: orgGroupIDs(), loginFails: true}

	err := keycloak.start(t).AddOrgMembers(context.Background(), "lab", map[string][]string{
		permissionValidation.OrgAdmin:  {"admin-1"},
		permissionValidation.OrgMember: {"member-1", "member-2"},
	})

	require.Error(t, err)
	for _, userID := range []string{"admin-1", "member-1", "member-2"} {
		assert.Contains(t, err.Error(), userID)
	}
	_, added := keycloak.writes()
	assert.Empty(t, added)
}
