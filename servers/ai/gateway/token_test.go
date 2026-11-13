package gateway

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenIssuer(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"https://keycloak.example/realms/prompt","sub":"u1"}`))
	assert.Equal(t, "https://keycloak.example/realms/prompt", tokenIssuer("Bearer header."+payload+".signature"))

	for _, header := range []string{"", "Bearer", "Bearer a.b", "Bearer a.!!!.c", "Bearer a." + base64.RawURLEncoding.EncodeToString([]byte(`[]`)) + ".c"} {
		assert.Empty(t, tokenIssuer(header), header)
	}
}
