package keycloakRealmManager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

func (s *KeycloakRealmService) DeleteOrgGroupsAndRoles(ctx context.Context, slug string) error {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return err
	}

	groupPath := ORGS_GROUP_NAME + "/" + slug
	groupID, found, err := s.findGroupAtPath(ctx, token.AccessToken, groupPath)
	if err != nil {
		return err
	}

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
