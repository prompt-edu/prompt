package keycloakRealmManager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Nerzal/gocloak/v14"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// fakeRealm is an in-memory Keycloak realm. It serves the admin endpoints that provision and delete
// the groups and roles of a course, records the route pattern of every request, and answers the
// failing pattern with a 500 or closes the connection of the dropping pattern without an answer.
type fakeRealm struct {
	mu        sync.Mutex
	groups    map[string]*fakeGroup // group ID → group
	roles     map[string]bool       // client role names
	userRoles map[string][]string
	nextID    int
	requests  []string
	failing   string
	dropping  string
}

type fakeGroup struct {
	id, name, parentID string
	members            []string
	roles              []string
}

// fakeGroupState is what a test compares of a group: its ID, members, and mapped roles.
type fakeGroupState struct {
	ID      string
	Members []string
	Roles   []string
}

func newFakeRealm() *fakeRealm {
	return &fakeRealm{groups: map[string]*fakeGroup{}, roles: map[string]bool{}, userRoles: map[string][]string{}}
}

// addGroup creates the group at groupPath with the given members, and every missing ancestor.
func (f *fakeRealm) addGroup(groupPath string, members ...string) string {
	parentID := ""
	for _, name := range strings.Split(groupPath, "/") {
		group := f.child(parentID, name)
		if group == nil {
			group = f.newGroup(name, parentID)
		}
		parentID = group.id
	}
	f.groups[parentID].members = members
	return parentID
}

// addRole creates a client role and maps it to the groups at the given paths.
func (f *fakeRealm) addRole(roleName string, groupPaths ...string) {
	f.roles[roleName] = true
	for _, groupPath := range groupPaths {
		group := f.groupAt(groupPath)
		group.roles = append(group.roles, roleName)
	}
}

func (f *fakeRealm) grantRole(roleName, userID string) {
	f.userRoles[userID] = append(f.userRoles[userID], roleName)
}

func (f *fakeRealm) directRoles(userID string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.userRoles[userID])
}

func (f *fakeRealm) newGroup(name, parentID string) *fakeGroup {
	f.nextID++
	group := &fakeGroup{id: fmt.Sprintf("group-%d", f.nextID), name: name, parentID: parentID}
	f.groups[group.id] = group
	return group
}

func (f *fakeRealm) child(parentID, name string) *fakeGroup {
	for _, group := range f.groups {
		if group.parentID == parentID && group.name == name {
			return group
		}
	}
	return nil
}

func (f *fakeRealm) groupAt(groupPath string) *fakeGroup {
	var group *fakeGroup
	parentID := ""
	for _, name := range strings.Split(strings.TrimPrefix(groupPath, "/"), "/") {
		if group = f.child(parentID, name); group == nil {
			return nil
		}
		parentID = group.id
	}
	return group
}

func (f *fakeRealm) pathOf(group *fakeGroup) string {
	if group.parentID == "" {
		return group.name
	}
	return f.pathOf(f.groups[group.parentID]) + "/" + group.name
}

func (f *fakeRealm) deleteGroup(id string) {
	for _, group := range f.groups {
		if group.parentID == id {
			f.deleteGroup(group.id)
		}
	}
	delete(f.groups, id)
}

// state returns every group by path, and the sorted client role names.
func (f *fakeRealm) state() (map[string]fakeGroupState, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	groups := map[string]fakeGroupState{}
	for _, group := range f.groups {
		groups[f.pathOf(group)] = fakeGroupState{ID: group.id, Members: slices.Clone(group.members), Roles: slices.Clone(group.roles)}
	}
	var roles []string
	for role := range f.roles {
		roles = append(roles, role)
	}
	slices.Sort(roles)
	return groups, roles
}

func (f *fakeRealm) recorded() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.requests)
}

func (f *fakeRealm) start(t *testing.T) *KeycloakRealmService {
	writeJSON := func(w http.ResponseWriter, status int, body any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}
	created := func(w http.ResponseWriter, location string) {
		w.Header().Set("Location", location)
		w.WriteHeader(http.StatusCreated)
	}
	notFound := func(w http.ResponseWriter) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	}

	mux := http.NewServeMux()
	handle := func(pattern string, handler http.HandlerFunc) {
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.requests = append(f.requests, pattern)
			switch pattern {
			case f.dropping:
				if conn, _, err := w.(http.Hijacker).Hijack(); err == nil {
					_ = conn.Close()
				}
				return
			case f.failing:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unavailable"})
				return
			}
			handler(w, r)
		})
	}

	handle("POST /realms/prompt/protocol/openid-connect/token", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"access_token": "token", "expires_in": 60})
	})
	handle("GET /admin/realms/prompt/group-by-path/{path...}", func(w http.ResponseWriter, r *http.Request) {
		group := f.groupAt(r.PathValue("path"))
		if group == nil {
			notFound(w)
			return
		}
		writeJSON(w, http.StatusOK, gocloak.Group{ID: &group.id, Name: &group.name})
	})
	createGroup := func(w http.ResponseWriter, r *http.Request, parentID string) {
		var body gocloak.Group
		_ = json.NewDecoder(r.Body).Decode(&body)
		if f.child(parentID, *body.Name) != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"errorMessage": "Top level group named '" + *body.Name + "' already exists."})
			return
		}
		created(w, "http://keycloak/admin/realms/prompt/groups/"+f.newGroup(*body.Name, parentID).id)
	}
	handle("POST /admin/realms/prompt/groups", func(w http.ResponseWriter, r *http.Request) {
		createGroup(w, r, "")
	})
	handle("POST /admin/realms/prompt/groups/{id}/children", func(w http.ResponseWriter, r *http.Request) {
		if f.groups[r.PathValue("id")] == nil {
			notFound(w)
			return
		}
		createGroup(w, r, r.PathValue("id"))
	})
	handle("DELETE /admin/realms/prompt/groups/{id}", func(w http.ResponseWriter, r *http.Request) {
		if f.groups[r.PathValue("id")] == nil {
			notFound(w)
			return
		}
		f.deleteGroup(r.PathValue("id"))
		w.WriteHeader(http.StatusNoContent)
	})
	handle("GET /admin/realms/prompt/groups/{id}/children", func(w http.ResponseWriter, r *http.Request) {
		if f.groups[r.PathValue("id")] == nil {
			notFound(w)
			return
		}
		children := []gocloak.Group{}
		for _, group := range f.groups {
			if group.parentID == r.PathValue("id") {
				children = append(children, gocloak.Group{ID: gocloak.StringP(group.id), Name: gocloak.StringP(group.name)})
			}
		}
		slices.SortFunc(children, func(a, b gocloak.Group) int { return strings.Compare(*a.Name, *b.Name) })
		first, _ := strconv.Atoi(r.URL.Query().Get("first"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("max"))
		writeJSON(w, http.StatusOK, children[min(first, len(children)):min(first+pageSize, len(children))])
	})
	handle("POST /admin/realms/prompt/groups/{id}/role-mappings/clients/{client}", func(w http.ResponseWriter, r *http.Request) {
		group := f.groups[r.PathValue("id")]
		if group == nil {
			notFound(w)
			return
		}
		var roles []gocloak.Role
		_ = json.NewDecoder(r.Body).Decode(&roles)
		for _, role := range roles {
			if !slices.Contains(group.roles, *role.Name) {
				group.roles = append(group.roles, *role.Name)
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handle("GET /admin/realms/prompt/clients/{client}/roles/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if !f.roles[name] {
			notFound(w)
			return
		}
		writeJSON(w, http.StatusOK, gocloak.Role{ID: gocloak.StringP("role-" + name), Name: &name})
	})
	handle("POST /admin/realms/prompt/clients/{client}/roles", func(w http.ResponseWriter, r *http.Request) {
		var role gocloak.Role
		_ = json.NewDecoder(r.Body).Decode(&role)
		if f.roles[*role.Name] {
			writeJSON(w, http.StatusConflict, map[string]string{"errorMessage": "Role with name " + *role.Name + " already exists"})
			return
		}
		f.roles[*role.Name] = true
		created(w, "http://keycloak/admin/realms/prompt/clients/"+r.PathValue("client")+"/roles/"+*role.Name)
	})
	handle("DELETE /admin/realms/prompt/clients/{client}/roles/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if !f.roles[name] {
			notFound(w)
			return
		}
		delete(f.roles, name)
		for _, group := range f.groups {
			group.roles = slices.DeleteFunc(group.roles, func(role string) bool { return role == name })
		}
		for userID, roles := range f.userRoles {
			f.userRoles[userID] = slices.DeleteFunc(roles, func(role string) bool { return role == name })
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handle("PUT /admin/realms/prompt/users/{user}/groups/{group}", func(w http.ResponseWriter, r *http.Request) {
		group := f.groups[r.PathValue("group")]
		if group == nil {
			notFound(w)
			return
		}
		if !slices.Contains(group.members, r.PathValue("user")) {
			group.members = append(group.members, r.PathValue("user"))
		}
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return NewKeycloakRealmService(server.URL, "prompt", "prompt-server", "secret", "client-uuid", db.Queries{})
}
