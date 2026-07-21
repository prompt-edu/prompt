package keycloakTokenVerifier

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

// Global verifier, initialized at application start-up
var verifier *oidc.IDTokenVerifier

func InitKeycloakVerifier() error {
	ctx := context.Background()

	// Construct the provider URL. Keycloak hosts OIDC metadata at:
	//   {BaseURL}/realms/{Realm}/.well-known/openid-configuration
	providerURL := fmt.Sprintf("%s/realms/%s", KeycloakTokenVerifierSingleton.BaseURL, KeycloakTokenVerifierSingleton.Realm)

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		return fmt.Errorf("failed to create new OIDC provider: %w", err)
	}

	// Configure the verifier with the expected client ID (audience)
	config := &oidc.Config{
		SkipClientIDCheck: true, // otherwise students cannot apply to courses
	}

	verifier = provider.Verifier(config)
	return nil
}
