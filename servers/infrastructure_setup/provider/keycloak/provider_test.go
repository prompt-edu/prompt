package keycloak

import "testing"

func TestKeycloakAuthFields(t *testing.T) {
	fields := (&Provider{}).GetAuthFields()
	if len(fields) != 4 {
		t.Fatalf("auth fields = %d, want 4", len(fields))
	}
	required := map[string]bool{}
	for _, field := range fields {
		required[field.Name] = field.Required
	}
	for _, name := range []string{"keycloak_url", "realm", "client_id", "client_secret"} {
		if !required[name] {
			t.Fatalf("field %q is not marked required in %+v", name, fields)
		}
	}
}

func TestNewStoresConfig(t *testing.T) {
	provider := New(Config{
		KeycloakURL:  "https://keycloak.example.com",
		Realm:        "prompt",
		ClientID:     "client",
		ClientSecret: "secret",
	})
	if provider.cfg.Realm != "prompt" {
		t.Fatalf("Realm = %q, want prompt", provider.cfg.Realm)
	}
}
