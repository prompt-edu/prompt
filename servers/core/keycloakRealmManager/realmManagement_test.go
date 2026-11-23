package keycloakRealmManager

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGroupMembersReturnsEveryMemberAcrossPages(t *testing.T) {
	for _, count := range []int{0, groupMembersPageSize - 1, groupMembersPageSize, 2*groupMembersPageSize + 1} {
		t.Run(fmt.Sprintf("%d members", count), func(t *testing.T) {
			userIDs := make([]string, count)
			for i := range userIDs {
				userIDs[i] = fmt.Sprintf("student-%03d", i)
			}
			keycloak := &fakeKeycloak{members: map[string][]string{"custom-group": userIDs}}

			members, err := keycloak.start(t).GetGroupMembers(context.Background(), "token", "custom-group")

			require.NoError(t, err)
			memberIDs := make([]string, len(members))
			for i, member := range members {
				memberIDs[i] = *member.ID
			}
			assert.Equal(t, userIDs, memberIDs)
		})
	}
}

func TestGetGroupMembersReturnsTheKeycloakError(t *testing.T) {
	keycloak := &fakeKeycloak{membersFail: true}

	members, err := keycloak.start(t).GetGroupMembers(context.Background(), "token", "custom-group")

	assert.Error(t, err)
	assert.Nil(t, members)
}

// unauthorizedKeycloak answers every request with a 401, like Keycloak 26.8.2 does on the
// Admin REST API when the service account lacks the kc-admin-api client scope.
func unauthorizedKeycloak(t *testing.T) *KeycloakRealmService {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"HTTP 401 Unauthorized"}`, http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)
	return NewKeycloakRealmService(server.URL, "prompt", "prompt-server", "secret", "client-uuid", db.Queries{})
}

func TestAdminAPIUnauthorizedErrorPointsToTheClientScopeDocs(t *testing.T) {
	_, err := unauthorizedKeycloak(t).GetGroupByPath(context.Background(), "token", "/Prompt/ios2526-iPraktikum/Lecturer", "Lecturer")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "kc-admin-api")
	assert.Contains(t, err.Error(), adminAPIScopeDocsURL)
}

func TestLoginUnauthorizedErrorDoesNotBlameTheClientScope(t *testing.T) {
	_, err := unauthorizedKeycloak(t).LoginClient(context.Background())

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "kc-admin-api")
}
