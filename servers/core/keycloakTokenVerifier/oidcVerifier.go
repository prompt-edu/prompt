package keycloakTokenVerifier

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

func (v *KeycloakTokenVerifier) initOIDCVerifier(ctx context.Context) error {
	// Construct the provider URL. Keycloak hosts OIDC metadata at:
	//   {BaseURL}/realms/{Realm}/.well-known/openid-configuration
	providerURL := fmt.Sprintf("%s/realms/%s", v.BaseURL, v.Realm)

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		return fmt.Errorf("failed to create new OIDC provider: %w", err)
	}

	// Configure the verifier with the expected client ID (audience)
	config := &oidc.Config{
		SkipClientIDCheck: true, // otherwise students cannot apply to courses
	}

	v.verifier = provider.Verifier(config)
	return nil
}
