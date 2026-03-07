package keycloakTokenVerifier

import (
	"context"
	"log"

	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type KeycloakTokenVerifier struct {
	BaseURL                 string
	Realm                   string
	ClientID                string
	expectedAuthorizedParty string
	queries                 db.Queries
}

var KeycloakTokenVerifierSingleton *KeycloakTokenVerifier

var TOP_LEVEL_GROUP_NAME = "Prompt"

func InitKeycloakTokenVerifier(ctx context.Context, BaseURL, Realm, ClientID, expectedAuthorizedParty string, queries db.Queries) {
	KeycloakTokenVerifierSingleton = &KeycloakTokenVerifier{
		BaseURL:                 BaseURL,
		Realm:                   Realm,
		ClientID:                ClientID,
		expectedAuthorizedParty: expectedAuthorizedParty,
		queries:                 queries,
	}

	// init the middleware
	err := InitKeycloakVerifier()
	if err != nil {
		log.Fatal("Failed to initialize Keycloak verifier: ", err)
	}
}
