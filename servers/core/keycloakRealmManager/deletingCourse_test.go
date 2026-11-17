package keycloakRealmManager

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsCourseClientRole(t *testing.T) {
	const group = "ws24-ios"
	cases := []struct {
		name     string
		roleName string
		want     bool
	}{
		{"lecturer role", group + "-Lecturer", true},
		{"editor role", group + "-Editor", true},
		{"custom group role", group + "-cg-teamA", true},
		{"unrelated role", "PromptAdmin", false},
		{"other course lecturer", "ws24-android-Lecturer", false},
		{"prefix course is not matched", "ws24-ios-advanced-Lecturer", false},
		{"prefix course custom role is not matched", "ws24-ios-advanced-cg-teamA", false},
		{"bare group name", group, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isCourseClientRole(tc.roleName, group); got != tc.want {
				t.Errorf("isCourseClientRole(%q, %q) = %v, want %v", tc.roleName, group, got, tc.want)
			}
		})
	}
}

// courseRealm holds the groups and roles of the course "lab" in the given semester, with a member in
// every group, next to another course's role.
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
	return realm
}

func TestDeleteCourseGroupsAndRolesByNameDeletesTheGroupLast(t *testing.T) {
	realm := courseRealm("ws24")

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws24")

	require.NoError(t, err)
	groups, roles := realm.state()
	assert.Equal(t, []string{"Prompt"}, slices.Sorted(maps.Keys(groups)))
	assert.Equal(t, []string{"ws24-lab-advanced-Lecturer"}, roles, "another course's roles stay")
	var deletes []string
	for _, request := range realm.recorded() {
		if strings.HasPrefix(request, "DELETE") {
			deletes = append(deletes, request)
		}
	}
	assert.Equal(t, []string{
		"DELETE /admin/realms/prompt/clients/{client}/roles/{name}",
		"DELETE /admin/realms/prompt/clients/{client}/roles/{name}",
		"DELETE /admin/realms/prompt/clients/{client}/roles/{name}",
		"DELETE /admin/realms/prompt/groups/{id}",
	}, deletes)
}

func TestDeleteCourseGroupsAndRolesByNameKeepsTheGroupWhenARoleCannotBeDeleted(t *testing.T) {
	realm := courseRealm("ws24")
	realm.failing = "DELETE /admin/realms/prompt/clients/{client}/roles/{name}"

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws24")

	require.Error(t, err)
	groups, _ := realm.state()
	assert.Equal(t, []string{"lecturer"}, groups["Prompt/ws24-lab/Lecturer"].Members, "the group keeps its members for a retry")
}

func TestDeleteCourseGroupsAndRolesByNameSucceedsWhenTheGroupIsAlreadyGone(t *testing.T) {
	realm := newFakeRealm()
	realm.addGroup("Prompt")
	realm.addRole("ws24-lab-Lecturer")

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "lab", "ws24")

	require.NoError(t, err)
	_, roles := realm.state()
	assert.Empty(t, roles)
}

// The path "Prompt/ws24-ios/Lecturer" resolves to the Lecturer group of the course "ws24-ios", which
// a course named "ios/Lecturer" must not delete.
func TestDeleteCourseGroupsAndRolesByNameLeavesAnotherCoursesGroupForASlashInTheName(t *testing.T) {
	realm := newFakeRealm()
	realm.addGroup("Prompt/ws24-ios/Lecturer", "lecturer")
	realm.addRole("ws24-ios-Lecturer", "Prompt/ws24-ios/Lecturer")

	err := realm.start(t).DeleteCourseGroupsAndRolesByName(context.Background(), "ios/Lecturer", "ws24")

	require.NoError(t, err)
	groups, roles := realm.state()
	assert.Equal(t, []string{"lecturer"}, groups["Prompt/ws24-ios/Lecturer"].Members)
	assert.Equal(t, []string{"ws24-ios-Lecturer"}, roles)
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
