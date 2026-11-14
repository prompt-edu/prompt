package org

import (
	"context"
	"slices"
	"sync"

	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

// fakeMemberDirectory keeps org memberships in memory, one role per user and org, as the Keycloak
// groups do.
type fakeMemberDirectory struct {
	mu    sync.Mutex
	users []keycloakRealmDTO.StaffMember
	roles map[string]map[string]string // org slug → Keycloak user ID → org role
	err   error                        // returned by every call when set
}

func newFakeMemberDirectory(users ...keycloakRealmDTO.StaffMember) *fakeMemberDirectory {
	return &fakeMemberDirectory{users: users, roles: map[string]map[string]string{}}
}

func staffMember(id, universityLogin string) keycloakRealmDTO.StaffMember {
	return keycloakRealmDTO.StaffMember{KeycloakUserID: id, Username: id, UniversityLogin: universityLogin}
}

func (d *fakeMemberDirectory) roleOf(slug, userID string) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.roles[slug][userID]
}

func (d *fakeMemberDirectory) GetOrgMembers(_ context.Context, slug string) ([]keycloakRealmDTO.OrgMember, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.err != nil {
		return nil, d.err
	}
	members := []keycloakRealmDTO.OrgMember{}
	for _, orgRole := range []string{permissionValidation.OrgAdmin, permissionValidation.OrgMember} {
		for _, user := range d.users {
			if d.roles[slug][user.KeycloakUserID] == orgRole {
				members = append(members, keycloakRealmDTO.OrgMember{StaffMember: user, Role: orgRole})
			}
		}
	}
	return members, nil
}

func (d *fakeMemberDirectory) FindUserByUniversityLogin(_ context.Context, universityLogin string) (keycloakRealmDTO.StaffMember, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.err != nil {
		return keycloakRealmDTO.StaffMember{}, d.err
	}
	index := slices.IndexFunc(d.users, func(user keycloakRealmDTO.StaffMember) bool { return user.UniversityLogin == universityLogin })
	if index < 0 {
		return keycloakRealmDTO.StaffMember{}, keycloakRealmManager.ErrUniversityLoginNotFound
	}
	return d.users[index], nil
}

// knows reports whether the user exists; like Keycloak, the writes fail for any other user.
func (d *fakeMemberDirectory) knows(userID string) bool {
	return slices.ContainsFunc(d.users, func(user keycloakRealmDTO.StaffMember) bool { return user.KeycloakUserID == userID })
}

func (d *fakeMemberDirectory) SetOrgMemberRole(_ context.Context, slug, userID, orgRole string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.err != nil {
		return d.err
	}
	if !d.knows(userID) {
		return keycloakRealmManager.ErrUserNotFound
	}
	if d.roles[slug] == nil {
		d.roles[slug] = map[string]string{}
	}
	d.roles[slug][userID] = orgRole
	return nil
}

func (d *fakeMemberDirectory) RemoveOrgMember(_ context.Context, slug, userID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.err != nil {
		return d.err
	}
	if !d.knows(userID) {
		return keycloakRealmManager.ErrUserNotFound
	}
	delete(d.roles[slug], userID)
	return nil
}
