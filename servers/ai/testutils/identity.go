package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	promptSDK "github.com/prompt-edu/prompt-sdk"
)

// Identity fakes Keycloak and core, so tests run the SDK's real AuthenticationMiddleware.
type Identity struct {
	CoreURL    string
	PhaseTypes map[string]string
	issuer     *httptest.Server
	core       *httptest.Server
	signer     jose.Signer
}

func StartIdentity() (*Identity, func(), error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: privateKey, KeyID: "test"}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return nil, nil, err
	}
	identity := &Identity{PhaseTypes: map[string]string{}, signer: signer}

	identity.issuer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issuerURL := identity.issuer.URL + "/realms/prompt"
		switch r.URL.Path {
		case "/realms/prompt/.well-known/openid-configuration":
			writeJSON(w, map[string]any{
				"issuer":                                issuerURL,
				"jwks_uri":                              issuerURL + "/certs",
				"authorization_endpoint":                issuerURL + "/auth",
				"token_endpoint":                        issuerURL + "/token",
				"id_token_signing_alg_values_supported": []string{"RS256"},
			})
		case "/realms/prompt/certs":
			writeJSON(w, jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
				{Key: &privateKey.PublicKey, KeyID: "test", Algorithm: "RS256", Use: "sig"},
			}})
		default:
			http.NotFound(w, r)
		}
	}))

	identity.core = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rest, ok := strings.CutPrefix(r.URL.Path, "/api/auth/course_phase/"); ok && strings.HasSuffix(rest, "/roles") {
			phaseID := strings.TrimSuffix(rest, "/roles")
			writeJSON(w, map[string]string{
				"courseLecturerRole": LecturerRole(phaseID),
				"courseEditorRole":   EditorRole(phaseID),
				"customRolePrefix":   phaseID + "-cg-",
			})
			return
		}
		if phaseID, ok := strings.CutPrefix(r.URL.Path, "/api/course_phases/"); ok {
			phaseType, known := identity.PhaseTypes[phaseID]
			if !known {
				phaseType = "Assessment"
			}
			writeJSON(w, map[string]string{"id": phaseID, "coursePhaseTypeName": phaseType})
			return
		}
		http.NotFound(w, r)
	}))
	identity.CoreURL = identity.core.URL

	stop := func() {
		identity.issuer.Close()
		identity.core.Close()
	}
	if err := promptSDK.InitAuthenticationMiddleware(identity.issuer.URL, "prompt", identity.core.URL); err != nil {
		stop()
		return nil, nil, fmt.Errorf("init the SDK against the test issuer: %w", err)
	}
	return identity, stop, nil
}

func LecturerRole(coursePhaseID string) string { return coursePhaseID + "-Lecturer" }

func EditorRole(coursePhaseID string) string { return coursePhaseID + "-Editor" }

func (i *Identity) Token(subject string, roles ...string) string {
	now := time.Now()
	claims, _ := json.Marshal(map[string]any{
		"iss":             i.issuer.URL + "/realms/prompt",
		"sub":             subject,
		"aud":             []string{"prompt-server"},
		"azp":             "prompt-client",
		"iat":             now.Unix(),
		"exp":             now.Add(time.Hour).Unix(),
		"resource_access": map[string]any{"prompt-server": map[string]any{"roles": roles}},
	})
	signed, err := i.signer.Sign(claims)
	if err != nil {
		panic(err)
	}
	token, err := signed.CompactSerialize()
	if err != nil {
		panic(err)
	}
	return "Bearer " + token
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
