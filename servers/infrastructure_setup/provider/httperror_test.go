package provider

import (
	"strings"
	"testing"
)

// The error message is persisted on the instance and rendered in the UI, so the upstream
// body must not appear in it.
func TestHTTPErrorOmitsResponseBody(t *testing.T) {
	body := []byte(`{"error":"invalid_client","secret":"super-secret-value"}`)

	err := HTTPError("keycloak", "POST", "token", 401, body)

	if strings.Contains(err.Error(), "super-secret-value") || strings.Contains(err.Error(), "invalid_client") {
		t.Fatalf("error leaks the response body: %v", err)
	}
	for _, want := range []string{"keycloak", "POST", "token", "401"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want it to mention %q", err, want)
		}
	}
}
