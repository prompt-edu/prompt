package keycloakRealmManager

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

type KeycloakRealmService struct {
	client                  *gocloak.GoCloak
	BaseURL                 string
	Realm                   string
	ClientID                string
	ClientSecret            string
	idOfClient              string
	expectedAuthorizedParty string
	queries                 db.Queries
}

var KeycloakRealmSingleton *KeycloakRealmService

var TOP_LEVEL_GROUP_NAME = "Prompt"
var CUSTOM_GROUPS_NAME = "CustomGroups"

func InitKeycloak(ctx context.Context, router *gin.RouterGroup, BaseURL, Realm, ClientID, ClientSecret, idOfClient, expectedAuthorizedParty string, queries db.Queries) error {
	KeycloakRealmSingleton = &KeycloakRealmService{
		client:                  gocloak.NewClient(BaseURL),
		BaseURL:                 BaseURL,
		Realm:                   Realm,
		ClientID:                ClientID,
		ClientSecret:            ClientSecret,
		idOfClient:              idOfClient,
		expectedAuthorizedParty: expectedAuthorizedParty,
		queries:                 queries,
	}

	// Test Login connection
	_, err := LoginClient(ctx)

	// setup router
	setupKeycloakRouter(router, keycloakTokenVerifier.KeycloakMiddleware, checkAccessControlByIDWrapper)

	return err
}

// initializes the handler func with CheckCoursePermissions
func checkAccessControlByIDWrapper(allowedRoles ...string) gin.HandlerFunc {
	return permissionValidation.CheckAccessControlByID(permissionValidation.CheckCoursePermission, "courseID", allowedRoles...)
}
