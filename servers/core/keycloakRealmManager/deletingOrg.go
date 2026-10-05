package keycloakRealmManager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// DeleteOrgGroupsAndRoles removes the Keycloak artifacts created by
// CreateOrgGroupsAndRoles: the org's two client roles and the group /Orgs/<slug>
// (Keycloak cascades its Admin and Member subgroups and their memberships). The group
// goes last, so a failure part way through keeps the group and its memberships, and
// CreateOrgGroupsAndRoles can restore the roles onto it. Anything that is already gone
// is treated as success, so the operation is safe to retry and to run as the cleanup of
// a partially failed creation.
func (s *KeycloakRealmService) DeleteOrgGroupsAndRoles(ctx context.Context, slug string) error {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return err
	}

	// Resolve the group before deleting anything, so an unexpected group at the path
	// aborts the deletion untouched.
	groupPath := ORGS_GROUP_NAME + "/" + slug
	groupID, found, err := s.findGroupAtPath(ctx, token.AccessToken, groupPath)
	if err != nil {
		return err
	}

	// Group deletion only drops the group-role mappings, not the role definitions.
	for _, orgRole := range []string{permissionValidation.OrgAdmin, permissionValidation.OrgMember} {
		roleName := permissionValidation.OrgRoleName(slug, orgRole)
		if err := s.client.DeleteClientRole(ctx, token.AccessToken, s.Realm, s.idOfClient, roleName); err != nil && !hasStatus(err, http.StatusNotFound) {
			return fmt.Errorf("failed to delete client role %s: %w", roleName, err)
		}
	}

	if !found {
		log.Warnf("org %s group not found in Keycloak; skipping group deletion", slug)
		return nil
	}
	if err := s.client.DeleteGroup(ctx, token.AccessToken, s.Realm, groupID); err != nil && !hasStatus(err, http.StatusNotFound) {
		return fmt.Errorf("failed to delete org group: %w", err)
	}
	return nil
}
