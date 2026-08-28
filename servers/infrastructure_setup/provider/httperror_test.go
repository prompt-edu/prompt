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

// A per-item reason is worth showing, but it is upstream-controlled text landing in a
// persisted, UI-rendered field, so it must be flattened and bounded.
func TestUpstreamReasonIsSanitised(t *testing.T) {
	if got := UpstreamReason("Invite email\nis   invalid\r\n"); got != "Invite email is invalid" {
		t.Fatalf("UpstreamReason = %q, want the reason collapsed onto one line", got)
	}
	if got := UpstreamReason("bad\x00\x07value"); got != "badvalue" {
		t.Fatalf("UpstreamReason = %q, want control characters removed", got)
	}
	long := UpstreamReason(strings.Repeat("x", 500))
	if len(long) > maxUpstreamReasonLength+3 {
		t.Fatalf("UpstreamReason length = %d, want it truncated", len(long))
	}
	if !strings.HasSuffix(long, "...") {
		t.Fatalf("UpstreamReason = %q, want a truncation marker", long[len(long)-10:])
	}
}
