package keycloakRealmManager

import (
	"context"
	"testing"

	"github.com/Nerzal/gocloak/v14"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keycloakUser(id, username, universityLogin string) gocloak.User {
	user := gocloak.User{ID: gocloak.StringP(id), Username: gocloak.StringP(username)}
	if universityLogin != "" {
		user.Attributes = map[string][]string{"university_login": {universityLogin}}
	}
	return user
}

func TestGetOrgMembersListsAdminsFirstAndEveryUserOnce(t *testing.T) {
	keycloak := &fakeKeycloak{
		groupIDs: orgGroupIDs(),
		members:  map[string][]string{"admin-group": {"ada"}, "member-group": {"max", "ada"}},
		users:    []gocloak.User{keycloakUser("ada", "ada", "ad12min"), keycloakUser("max", "max", "ma34mbr")},
	}

	members, err := keycloak.start(t).GetOrgMembers(context.Background(), "lab")

	require.NoError(t, err)
	require.Len(t, members, 2)
	assert.Equal(t, "ada", members[0].KeycloakUserID)
	assert.Equal(t, "ad12min", members[0].UniversityLogin)
	assert.Equal(t, permissionValidation.OrgAdmin, members[0].Role)
	assert.Equal(t, "max", members[1].KeycloakUserID)
	assert.Equal(t, permissionValidation.OrgMember, members[1].Role)
}

func TestGetOrgMembersTreatsAMissingGroupAsEmpty(t *testing.T) {
	groupIDs := orgGroupIDs()
	delete(groupIDs, "Orgs/lab/Admin")
	keycloak := &fakeKeycloak{groupIDs: groupIDs, members: map[string][]string{"member-group": {"max"}}}

	members, err := keycloak.start(t).GetOrgMembers(context.Background(), "lab")

	require.NoError(t, err)
	require.Len(t, members, 1)
	assert.Equal(t, permissionValidation.OrgMember, members[0].Role)
}

func TestGetOrgMembersOfAnOrgWithoutMembersIsAnEmptyList(t *testing.T) {
	keycloak := &fakeKeycloak{groupIDs: orgGroupIDs()}

	members, err := keycloak.start(t).GetOrgMembers(context.Background(), "lab")

	require.NoError(t, err)
	assert.NotNil(t, members)
	assert.Empty(t, members)
}

func TestSetOrgMemberRoleAddsTheNewRoleBeforeRemovingTheOther(t *testing.T) {
	keycloak := &fakeKeycloak{groupIDs: orgGroupIDs()}

	err := keycloak.start(t).SetOrgMemberRole(context.Background(), "lab", "ada", permissionValidation.OrgAdmin)

	require.NoError(t, err)
	assert.Equal(t, []string{"PUT users/ada/groups/admin-group", "DELETE users/ada/groups/member-group"}, keycloak.writeLog)
}

func TestSetOrgMemberRoleToMember(t *testing.T) {
	keycloak := &fakeKeycloak{groupIDs: orgGroupIDs()}

	err := keycloak.start(t).SetOrgMemberRole(context.Background(), "lab", "ada", permissionValidation.OrgMember)

	require.NoError(t, err)
	assert.Equal(t, []string{"PUT users/ada/groups/member-group", "DELETE users/ada/groups/admin-group"}, keycloak.writeLog)
}

func TestSetOrgMemberRoleRejectsOtherRolesWithoutWriting(t *testing.T) {
	for _, role := range []string{"", "admin", permissionValidation.CourseLecturer, permissionValidation.PromptAdmin} {
		keycloak := &fakeKeycloak{groupIDs: orgGroupIDs()}

		err := keycloak.start(t).SetOrgMemberRole(context.Background(), "lab", "ada", role)

		assert.ErrorIs(t, err, ErrInvalidOrgRole, role)
		assert.Empty(t, keycloak.writeLog, role)
	}
}

func TestSetOrgMemberRoleOfAnUnknownUserRemovesNothing(t *testing.T) {
	keycloak := &fakeKeycloak{groupIDs: orgGroupIDs(), unknownUsers: map[string]bool{"ghost": true}}

	err := keycloak.start(t).SetOrgMemberRole(context.Background(), "lab", "ghost", permissionValidation.OrgAdmin)

	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Equal(t, []string{"PUT users/ghost/groups/admin-group"}, keycloak.writeLog)
}

func TestSetOrgMemberRoleWritesNothingWhenAGroupIsMissing(t *testing.T) {
	groupIDs := orgGroupIDs()
	delete(groupIDs, "Orgs/lab/Member")
	keycloak := &fakeKeycloak{groupIDs: groupIDs}

	err := keycloak.start(t).SetOrgMemberRole(context.Background(), "lab", "ada", permissionValidation.OrgAdmin)

	assert.ErrorContains(t, err, "the Member group of org lab is missing")
	assert.Empty(t, keycloak.writeLog)
}

func TestRemoveOrgMemberRemovesTheUserFromBothGroups(t *testing.T) {
	keycloak := &fakeKeycloak{groupIDs: orgGroupIDs()}

	err := keycloak.start(t).RemoveOrgMember(context.Background(), "lab", "ada")

	require.NoError(t, err)
	assert.Equal(t, []string{"DELETE users/ada/groups/admin-group", "DELETE users/ada/groups/member-group"}, keycloak.writeLog)
}

func TestRemoveOrgMemberSkipsAMissingGroup(t *testing.T) {
	groupIDs := orgGroupIDs()
	delete(groupIDs, "Orgs/lab/Admin")
	keycloak := &fakeKeycloak{groupIDs: groupIDs}

	err := keycloak.start(t).RemoveOrgMember(context.Background(), "lab", "ada")

	require.NoError(t, err)
	assert.Equal(t, []string{"DELETE users/ada/groups/member-group"}, keycloak.writeLog)
}

func TestRemoveOrgMemberOfAnUnknownUser(t *testing.T) {
	keycloak := &fakeKeycloak{groupIDs: orgGroupIDs(), unknownUsers: map[string]bool{"ghost": true}}

	err := keycloak.start(t).RemoveOrgMember(context.Background(), "lab", "ghost")

	assert.ErrorIs(t, err, ErrUserNotFound)
}
