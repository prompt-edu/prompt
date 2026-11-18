package keycloakRealmManager

import (
	"context"
	"fmt"
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

func TestCreateCourseGroupsAndRolesEndsInTheSameStateTheSecondTime(t *testing.T) {
	realm := newFakeRealm()
	service := realm.start(t)
	require.NoError(t, service.CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator"))
	groupsBefore, rolesBefore := realm.state()

	err := service.CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator")

	require.NoError(t, err)
	groupsAfter, rolesAfter := realm.state()
	assert.Equal(t, withoutIDs(groupsBefore), withoutIDs(groupsAfter))
	assert.Equal(t, rolesBefore, rolesAfter)
}

// A course group can be left over from a creation whose commit failed, or from a course deleted
// before deletions cleaned up Keycloak. Nobody may keep access to the new course through it, also
// not through a leftover role granted outside the course group.
func TestCreateCourseGroupsAndRolesDeletesTheLeftoversOfAnEarlierCourse(t *testing.T) {
	realm := newFakeRealm()
	leftoverGroupID := realm.addGroup("Prompt/ws24-ios")
	realm.addGroup("Prompt/ws24-ios/Lecturer", "old-lecturer")
	realm.addGroup("Prompt/ws24-ios/Editor", "old-editor")
	realm.addGroup("Prompt/ws24-ios/CustomGroups/team-a", "old-student")
	realm.addGroup("Other/staff", "outsider")
	realm.addRole("ws24-ios-Lecturer", "Prompt/ws24-ios/Lecturer", "Other/staff")
	realm.addRole("ws24-ios-Editor", "Prompt/ws24-ios/Editor")
	realm.addRole("ws24-ios-cg-team-a", "Prompt/ws24-ios/CustomGroups/team-a")
	realm.grantRole("ws24-ios-Lecturer", "student")
	realm.grantRole("ws24-ios-cg-team-a", "old-student")
	realm.addGroup("Prompt/ws24-ios-cg-team/Lecturer", "other-lecturer")
	realm.addRole("ws24-ios-cg-team-Lecturer", "Prompt/ws24-ios-cg-team/Lecturer")

	err := realm.start(t).CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator")

	require.NoError(t, err)
	groups, roles := realm.state()
	assert.NotEqual(t, leftoverGroupID, groups["Prompt/ws24-ios"].ID, "the course group is created afresh")
	assert.Equal(t, []string{"creator"}, groups["Prompt/ws24-ios/Lecturer"].Members)
	assert.Empty(t, groups["Prompt/ws24-ios/Editor"].Members)
	assert.NotContains(t, groups, "Prompt/ws24-ios/CustomGroups")
	assert.Empty(t, groups["Other/staff"].Roles)
	assert.Empty(t, realm.directRoles("student"))
	assert.Empty(t, realm.directRoles("old-student"))
	assert.Equal(t, []string{"ws24-ios-Editor", "ws24-ios-Lecturer", "ws24-ios-cg-team-Lecturer"}, roles, "the Lecturer role of the course \"ios-cg-team\" stays")
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
	scenarios := map[string]func() *fakeRealm{
		"empty realm": newFakeRealm,
		"leftover course": func() *fakeRealm {
			realm := newFakeRealm()
			realm.addGroup("Prompt/ws24-ios/Lecturer", "old-lecturer")
			realm.addGroup("Prompt/ws24-ios/Editor", "old-editor")
			realm.addGroup("Prompt/ws24-ios/CustomGroups/team-a", "old-student")
			realm.addRole("ws24-ios-Lecturer", "Prompt/ws24-ios/Lecturer")
			realm.addRole("ws24-ios-Editor", "Prompt/ws24-ios/Editor")
			realm.addRole("ws24-ios-cg-team-a", "Prompt/ws24-ios/CustomGroups/team-a")
			return realm
		},
	}
	for scenario, newRealm := range scenarios {
		succeeding := newRealm()
		require.NoError(t, succeeding.start(t).CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator"))
		for i, call := range succeeding.recordedCalls() {
			t.Run(fmt.Sprintf("%s/%02d %s", scenario, i+1, call), func(t *testing.T) {
				realm := newRealm()
				realm.failingRequest = i + 1

				err := realm.start(t).CreateCourseGroupsAndRoles(context.Background(), "ios", "ws24", "creator")

				var apiErr *gocloak.APIError
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, http.StatusInternalServerError, apiErr.Code)
				assert.Len(t, realm.recordedCalls(), i+1, "provisioning stops at the failing request")
			})
		}
	}
}

func withoutIDs(groups map[string]fakeGroupState) map[string]fakeGroupState {
	cleared := map[string]fakeGroupState{}
	for groupPath, group := range groups {
		group.ID = ""
		cleared[groupPath] = group
	}
	return cleared
}
