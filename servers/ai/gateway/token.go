package gateway

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

// tokenIssuer reads the iss claim of a bearer token the SDK middleware has already verified.
func tokenIssuer(authHeader string) string {
	token, _ := strings.CutPrefix(authHeader, "Bearer ")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Issuer string `json:"iss"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return ""
	}
	return claims.Issuer
}
