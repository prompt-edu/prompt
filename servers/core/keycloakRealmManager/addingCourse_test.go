package keycloakRealmManager

import (
	"context"
	"maps"
	"net/http"
	"slices"
	"testing"

	"github.com/Nerzal/gocloak/v14"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCourseGroupsAndRolesProvisionsANewCourse(t *testing.T) {
	realm := newFakeRealm()

	err := realm.start(t).CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator")

	require.NoError(t, err)
	groups, roles := realm.state()
	assert.Equal(t, []string{"Prompt", "Prompt/ws24-ios", "Prompt/ws24-ios/Editor", "Prompt/ws24-ios/Lecturer"}, slices.Sorted(maps.Keys(groups)))
	assert.Equal(t, []string{"creator"}, groups["Prompt/ws24-ios/Lecturer"].Members)
	assert.Equal(t, []string{"ws24-ios-Lecturer"}, groups["Prompt/ws24-ios/Lecturer"].Roles)
	assert.Empty(t, groups["Prompt/ws24-ios/Editor"].Members)
	assert.Equal(t, []string{"ws24-ios-Editor"}, groups["Prompt/ws24-ios/Editor"].Roles)
	assert.Equal(t, []string{"ws24-ios-Editor", "ws24-ios-Lecturer"}, roles)
}

func TestCreateCourseGroupsAndRolesChangesNothingTheSecondTime(t *testing.T) {
	realm := newFakeRealm()
	service := realm.start(t)
	require.NoError(t, service.CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator"))
	groupsBefore, rolesBefore := realm.state()
	requestsBefore := len(realm.recorded())

	err := service.CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator")

	require.NoError(t, err)
	groupsAfter, rolesAfter := realm.state()
	assert.Equal(t, groupsBefore, groupsAfter)
	assert.Equal(t, rolesBefore, rolesAfter)
	creatingOrDeleting := []string{
		"POST /admin/realms/prompt/groups",
		"POST /admin/realms/prompt/groups/{id}/children",
		"POST /admin/realms/prompt/clients/{client}/roles",
		"DELETE /admin/realms/prompt/groups/{id}",
		"DELETE /admin/realms/prompt/clients/{client}/roles/{name}",
		"DELETE /admin/realms/prompt/users/{user}/groups/{group}",
	}
	for _, request := range realm.recorded()[requestsBefore:] {
		assert.NotContains(t, creatingOrDeleting, request, "the second run must neither create nor delete anything")
	}
}

// A course group can be left over from a creation whose commit failed, or from a course deleted
// before deletions cleaned up Keycloak. None of its members may keep access to the new course.
func TestCreateCourseGroupsAndRolesResetsALeftoverCourseGroup(t *testing.T) {
	realm := newFakeRealm()
	courseGroupID := realm.addGroup("Prompt/ws24-ios")
	realm.addGroup("Prompt/ws24-ios/Lecturer", "old-lecturer")
	realm.addGroup("Prompt/ws24-ios/Editor", "old-editor")
	realm.addGroup("Prompt/ws24-ios/CustomGroups/team-a", "old-student")
	realm.addRole("ws24-ios-Lecturer", "Prompt/ws24-ios/Lecturer")
	realm.addRole("ws24-ios-Editor", "Prompt/ws24-ios/Editor")
	realm.addRole("ws24-ios-cg-team-a", "Prompt/ws24-ios/CustomGroups/team-a")

	err := realm.start(t).CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator")

	require.NoError(t, err)
	groups, _ := realm.state()
	assert.Equal(t, courseGroupID, groups["Prompt/ws24-ios"].ID, "the leftover course group is reused")
	assert.Equal(t, []string{"creator"}, groups["Prompt/ws24-ios/Lecturer"].Members)
	assert.Empty(t, groups["Prompt/ws24-ios/Editor"].Members)
	assert.NotContains(t, groups, "Prompt/ws24-ios/CustomGroups")
	assert.NotContains(t, groups, "Prompt/ws24-ios/CustomGroups/team-a")
}

// Keycloak's group search can return the top-level parents of matching subgroups, so the top-level
// group is looked up by its path instead.
func TestCreateCourseGroupsAndRolesFindsThePromptGroupByPath(t *testing.T) {
	realm := newFakeRealm()
	promptGroupID := realm.addGroup("Prompt")
	realm.addGroup("Other/Prompt")

	err := realm.start(t).CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator")

	require.NoError(t, err)
	assert.NotContains(t, realm.recorded(), "POST /admin/realms/prompt/groups")
	assert.Equal(t, promptGroupID, realm.groupAt("Prompt/ws24-ios").parentID)
}

func TestCreateCourseGroupsAndRolesKeepsTheCauseOfEveryFailure(t *testing.T) {
	steps := []struct {
		pattern  string
		leftover bool
	}{
		{pattern: "POST /realms/prompt/protocol/openid-connect/token"},
		{pattern: "GET /admin/realms/prompt/group-by-path/{path...}"},
		{pattern: "POST /admin/realms/prompt/groups"},
		{pattern: "POST /admin/realms/prompt/groups/{id}/children"},
		{pattern: "GET /admin/realms/prompt/clients/{client}/roles/{name}"},
		{pattern: "POST /admin/realms/prompt/clients/{client}/roles"},
		{pattern: "POST /admin/realms/prompt/groups/{id}/role-mappings/clients/{client}"},
		{pattern: "GET /admin/realms/prompt/groups/{id}/members"},
		{pattern: "PUT /admin/realms/prompt/users/{user}/groups/{group}"},
		{pattern: "DELETE /admin/realms/prompt/users/{user}/groups/{group}", leftover: true},
		{pattern: "DELETE /admin/realms/prompt/groups/{id}", leftover: true},
	}
	for _, step := range steps {
		t.Run(step.pattern, func(t *testing.T) {
			realm := newFakeRealm()
			if step.leftover {
				realm.addGroup("Prompt/ws24-ios/Lecturer", "old-lecturer")
				realm.addGroup("Prompt/ws24-ios/CustomGroups")
			}
			realm.failing = step.pattern

			err := realm.start(t).CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator")

			var apiErr *gocloak.APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, http.StatusInternalServerError, apiErr.Code)
			assert.Contains(t, realm.recorded(), step.pattern, "the failing step must have been reached")
		})
	}
}
