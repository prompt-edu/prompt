package keycloakTokenVerifier

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type KeycloakTokenVerifier struct {
	BaseURL                 string
	Realm                   string
	ClientID                string
	expectedAuthorizedParty string
	queries                 db.Queries
	verifier                *oidc.IDTokenVerifier
}

var TOP_LEVEL_GROUP_NAME = "Prompt"

func NewKeycloakTokenVerifier(ctx context.Context, BaseURL, Realm, ClientID, expectedAuthorizedParty string, queries db.Queries) (*KeycloakTokenVerifier, error) {
	tokenVerifier := &KeycloakTokenVerifier{
		BaseURL:                 BaseURL,
		Realm:                   Realm,
		ClientID:                ClientID,
		expectedAuthorizedParty: expectedAuthorizedParty,
		queries:                 queries,
	}

	if err := tokenVerifier.initOIDCVerifier(ctx); err != nil {
		return nil, err
	}

	return tokenVerifier, nil
}
