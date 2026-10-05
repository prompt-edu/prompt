package keycloakRealmManager

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Nerzal/gocloak/v14"
)

func TestHasStatus(t *testing.T) {
	// gocloak returns *APIError behind the error interface.
	var notFound error = &gocloak.APIError{Code: http.StatusNotFound}
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"not found response", &gocloak.APIError{Code: http.StatusNotFound, Message: "404 Not Found: Group path does not exist"}, true},
		{"wrapped not found response", fmt.Errorf("lookup: %w", notFound), true},
		{"status text without a code", &gocloak.APIError{Message: "could not get group: 404 Not Found"}, true},
		{"other status", &gocloak.APIError{Code: http.StatusForbidden, Message: "403 Forbidden"}, false},
		{"transport error with 404 in the URL", &gocloak.APIError{
			Message: `could not get group: Get "https://kc/admin/realms/prompt/group-by-path/Orgs/lab404": dial tcp: connection refused`,
		}, false},
		{"plain error", errors.New("404 Not Found"), false},
		{"no error", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasStatus(tc.err, http.StatusNotFound); got != tc.want {
				t.Errorf("hasStatus(%v, 404) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
