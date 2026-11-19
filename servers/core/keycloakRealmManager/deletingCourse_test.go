package keycloakRealmManager

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// courseRealm holds the groups and roles of the course "lab" in the given semester, with a member in
// every group, next to the roles of the courses "lab-advanced" and "lab-cg-team". Their names start
// like the staff and custom group roles of "lab".
func courseRealm(semester string) *fakeRealm {
	realm := newFakeRealm()
	course := "Prompt/" + semester + "-lab"
	realm.addGroup(course+"/Lecturer", "lecturer")
	realm.addGroup(course+"/Editor", "editor")
	realm.addGroup(course+"/CustomGroups/team-a", "student")
	realm.addRole(semester+"-lab-Lecturer", course+"/Lecturer")
	realm.addRole(semester+"-lab-Editor", course+"/Editor")
	realm.addRole(semester+"-lab-cg-team-a", course+"/CustomGroups/team-a")
	realm.addRole(semester + "-lab-advanced-Lecturer")
	realm.addGroup("Prompt/"+semester+"-lab-cg-team/Lecturer", "other-lecturer")
	realm.addRole(semester+"-lab-cg-team-Lecturer", "Prompt/"+semester+"-lab-cg-team/Lecturer")
	return realm
}

func deleteCalls(realm *fakeRealm) []string {
	var deletes []string
	for _, call := range realm.recordedCalls() {
		if strings.HasPrefix(call, "DELETE ") {
			deletes = append(deletes, call)
		}
	}
	return deletes
}

func TestDeleteCourseGroupsAndRolesByNameDeletesTheCustomRolesThenTheGroupThenTheStaffRoles(t *testing.T) {
	realm := courseRealm("ws24")
	courseGroupID := realm.groupAt("Prompt/ws24-lab").id

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws24")

	require.NoError(t, err)
	groups, roles := realm.state()
	assert.Equal(t, []string{"Prompt", "Prompt/ws24-lab-cg-team", "Prompt/ws24-lab-cg-team/Lecturer"}, slices.Sorted(maps.Keys(groups)))
	assert.Equal(t, []string{"ws24-lab-advanced-Lecturer", "ws24-lab-cg-team-Lecturer"}, roles, "other courses' roles stay")
	assert.Equal(t, []string{
		"DELETE /admin/realms/prompt/clients/client-uuid/roles/ws24-lab-cg-team-a",
		"DELETE /admin/realms/prompt/groups/" + courseGroupID,
		"DELETE /admin/realms/prompt/clients/client-uuid/roles/ws24-lab-Editor",
		"DELETE /admin/realms/prompt/clients/client-uuid/roles/ws24-lab-Lecturer",
	}, deleteCalls(realm))
}

func TestDeleteCourseGroupsAndRolesByNameDeletesTheRoleOfEveryCustomGroup(t *testing.T) {
	realm := newFakeRealm()
	for i := range childGroupsPageSize + 1 {
		team := fmt.Sprintf("Prompt/ws24-lab/CustomGroups/team-%03d", i)
		realm.addGroup(team)
		realm.addRole(fmt.Sprintf("ws24-lab-cg-team-%03d", i), team)
	}

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws24")

	require.NoError(t, err)
	_, roles := realm.state()
	assert.Empty(t, roles)
}

func TestDeleteCourseGroupsAndRolesByNameKeepsTheGroupWhenACustomRoleCannotBeDeleted(t *testing.T) {
	realm := courseRealm("ws24")
	realm.failing = "DELETE /admin/realms/prompt/clients/{client}/roles/{name}"

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws24")

	require.Error(t, err)
	groups, _ := realm.state()
	assert.Equal(t, []string{"lecturer"}, groups["Prompt/ws24-lab/Lecturer"].Members, "the group keeps its members for a retry")
}

func TestDeleteCourseGroupsAndRolesByNameKeepsTheLecturerRoleWhenTheGroupCannotBeDeleted(t *testing.T) {
	realm := courseRealm("ws24")
	realm.failing = "DELETE /admin/realms/prompt/groups/{id}"

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws24")

	require.Error(t, err)
	groups, roles := realm.state()
	assert.Equal(t, []string{"lecturer"}, groups["Prompt/ws24-lab/Lecturer"].Members)
	assert.Equal(t, []string{"ws24-lab-Lecturer"}, groups["Prompt/ws24-lab/Lecturer"].Roles, "the Lecturer keeps the role to retry the deletion")
	assert.Contains(t, roles, "ws24-lab-Lecturer")
}

func TestDeleteCourseGroupsAndRolesByNameSucceedsWhenTheGroupIsAlreadyGone(t *testing.T) {
	realm := newFakeRealm()
	realm.addGroup("Prompt")
	realm.addRole("ws24-lab-Lecturer")
	realm.addRole("ws24-lab-cg-team-Lecturer")

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws24")

	require.NoError(t, err)
	_, roles := realm.state()
	assert.Equal(t, []string{"ws24-lab-cg-team-Lecturer"}, roles, "the Lecturer role of the course \"lab-cg-team\" stays")
	assert.Equal(t, []string{"DELETE /admin/realms/prompt/clients/client-uuid/roles/ws24-lab-Lecturer"}, deleteCalls(realm), "the missing Editor role is not deleted")
}

// A "/" splits the group path, a "%" makes the URL invalid, and a "?", "#" or ";" cuts it short, so
// the course "lab#1" would look up the group and roles of the course "lab".
func TestDeleteCourseGroupsAndRolesByNameLeavesKeycloakAloneForAnUnaddressableName(t *testing.T) {
	for _, name := range []string{"lab/Lecturer", "lab%1", "lab?1", "lab#1", "lab;1", "lab\t1"} {
		t.Run(name, func(t *testing.T) {
			realm := courseRealm("ws24")
			groupsBefore, rolesBefore := realm.state()

			err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), name, "ws24")

			require.NoError(t, err)
			assert.Empty(t, realm.recorded())
			groupsAfter, rolesAfter := realm.state()
			assert.Equal(t, groupsBefore, groupsAfter)
			assert.Equal(t, rolesBefore, rolesAfter)
		})
	}
}

func TestDeleteCourseGroupsAndRolesByNameKeepsARoleAGroupOfAnotherCourseMaps(t *testing.T) {
	cases := []struct{ deleted, otherCourseGroup string }{
		{deleted: "lab", otherCourseGroup: "Prompt/ws24-lab-cg-team/Lecturer"},
		{deleted: "lab-cg-team", otherCourseGroup: "Prompt/ws24-lab/CustomGroups/team-Lecturer"},
	}
	for _, tc := range cases {
		t.Run(tc.deleted, func(t *testing.T) {
			realm := courseRealm("ws24")
			realm.addGroup("Prompt/ws24-lab/CustomGroups/team-Lecturer", "student")
			realm.addRole("ws24-lab-cg-team-Lecturer", "Prompt/ws24-lab/CustomGroups/team-Lecturer")

			err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), tc.deleted, "ws24")

			require.NoError(t, err)
			groups, roles := realm.state()
			assert.NotContains(t, groups, "Prompt/ws24-"+tc.deleted)
			assert.Contains(t, roles, "ws24-lab-cg-team-Lecturer", "the custom group role of \"lab\" is the Lecturer role of \"lab-cg-team\"")
			assert.Equal(t, []string{"ws24-lab-cg-team-Lecturer"}, groups[tc.otherCourseGroup].Roles)
		})
	}
}

func TestDeleteCourseGroupsAndRolesByNameSkipsAnUnaddressableCustomRole(t *testing.T) {
	realm := courseRealm("ws24")
	realm.addGroup("Prompt/ws24-lab/CustomGroups/team-Lecturer?x")

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws24")

	require.NoError(t, err)
	for _, call := range realm.recordedCalls() {
		assert.NotContains(t, call, "/roles/ws24-lab-cg-team-Lecturer", "the role of \"team-Lecturer?x\" must not be looked up as the Lecturer role of the course \"lab-cg-team\"")
	}
	_, roles := realm.state()
	assert.Contains(t, roles, "ws24-lab-cg-team-Lecturer")
}

// A transport error embeds the request URL, here one with "404" in the course identifier. It must
// not read as "the group is already gone".
func TestDeleteCourseGroupsAndRolesByNameFailsWhenKeycloakIsUnreachable(t *testing.T) {
	realm := courseRealm("ws404")
	realm.dropping = "GET /admin/realms/prompt/group-by-path/{path...}"
	groupsBefore, rolesBefore := realm.state()

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws404")

	require.ErrorContains(t, err, "group-by-path/Prompt/ws404-lab", "the error carries the URL the old text match found \"404\" in")
	groupsAfter, rolesAfter := realm.state()
	assert.Equal(t, groupsBefore, groupsAfter)
	assert.Equal(t, rolesBefore, rolesAfter)
}
