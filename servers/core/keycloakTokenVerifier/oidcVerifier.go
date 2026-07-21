package keycloakTokenVerifier

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	log "github.com/sirupsen/logrus"
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

	// Local-development escape hatch: when the browser reaches Keycloak on a
	// different host/port than the containerized server (e.g. http://127.0.0.1
	// vs the internal service name), the token issuer will not match the
	// server's provider URL. Signatures are still verified against the realm's
	// JWKS. NEVER enable this in production.
	if os.Getenv("KEYCLOAK_INSECURE_SKIP_ISSUER_CHECK") == "true" {
		config.SkipIssuerCheck = true
		log.Warn("KEYCLOAK_INSECURE_SKIP_ISSUER_CHECK is enabled - token issuer validation is disabled. Do not use this in production.")
	}

	verifier = provider.Verifier(config)
	return nil
}
