package keycloakRealmManager

import (
	"context"
	"errors"
	"testing"

	"github.com/Nerzal/gocloak/v14"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKeycloakRealmService() *KeycloakRealmService {
	return NewKeycloakRealmService("http://localhost:8081", "prompt", "prompt-server", "", "id-of-client", db.Queries{})
}

func TestIsAllowedCourseGroup(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"lecturer constant", permissionValidation.CourseLecturer, true},
		{"editor constant", permissionValidation.CourseEditor, true},
		{"empty", "", false},
		{"student", permissionValidation.CourseStudent, false},
		{"admin", "Admin", false},
		{"path traversal attempt", "../Lecturer", false},
		{"case sensitivity", "lecturer", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAllowedCourseGroup(tc.input); got != tc.want {
				t.Errorf("isAllowedCourseGroup(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// These three tests cover the pure-logic early returns that fire before any
// Keycloak interaction. End-to-end Keycloak behaviour is covered by the manual
// verification step in the plan.

func TestAddUserToCourseGroup_RejectsBadGroupName(t *testing.T) {
	err := testKeycloakRealmService().AddUserToCourseGroup(context.Background(), uuid.New(), "Admin", "user-1", "caller-1")
	if !errors.Is(err, ErrInvalidGroupName) {
		t.Errorf("expected ErrInvalidGroupName, got %v", err)
	}
}

func TestRemoveUserFromCourseGroup_RejectsBadGroupName(t *testing.T) {
	err := testKeycloakRealmService().RemoveUserFromCourseGroup(context.Background(), uuid.New(), "Admin", "user-1", "caller-1")
	if !errors.Is(err, ErrInvalidGroupName) {
		t.Errorf("expected ErrInvalidGroupName, got %v", err)
	}
}

func TestRemoveUserFromCourseGroup_RejectsSelfRemoval(t *testing.T) {
	const userID = "abc-123"
	err := testKeycloakRealmService().RemoveUserFromCourseGroup(context.Background(), uuid.New(), permissionValidation.CourseLecturer, userID, userID)
	if !errors.Is(err, ErrSelfRemoval) {
		t.Errorf("expected ErrSelfRemoval, got %v", err)
	}
}

func TestSearchKeycloakUsers_RejectsShortQuery(t *testing.T) {
	cases := []string{"", " ", "a", "  x  "}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			_, err := testKeycloakRealmService().SearchKeycloakUsers(context.Background(), q, 20)
			if !errors.Is(err, ErrInvalidQuery) {
				t.Errorf("expected ErrInvalidQuery for %q, got %v", q, err)
			}
		})
	}
}

func resultIDs(results keycloakRealmDTO.UserSearchResults) []string {
	ids := []string{}
	for _, user := range results.Results {
		ids = append(ids, user.KeycloakUserID)
	}
	return ids
}

// Keycloak's search matches bob's username; ada's university login is the query, and eve's only
// starts with it.
func TestSearchKeycloakUsersListsTheUserWithTheUniversityLoginFirst(t *testing.T) {
	keycloak := &fakeKeycloak{users: []gocloak.User{
		keycloakUser("bob", "ab12cde.old", "bo34bbb"),
		keycloakUser("ada", "ada", "ab12cde"),
		keycloakUser("eve", "eve", "ab12cdef"),
	}}

	results, err := keycloak.start(t).SearchKeycloakUsers(context.Background(), "AB12CDE", 20)

	require.NoError(t, err)
	assert.Equal(t, []string{"ada", "bob"}, resultIDs(results))
	assert.False(t, results.Truncated)
}

func TestSearchKeycloakUsersListsAUserMatchedTwiceOnce(t *testing.T) {
	keycloak := &fakeKeycloak{users: []gocloak.User{keycloakUser("ada", "ab12cde", "ab12cde")}}

	results, err := keycloak.start(t).SearchKeycloakUsers(context.Background(), "ab12cde", 20)

	require.NoError(t, err)
	assert.Equal(t, []string{"ada"}, resultIDs(results))
}

func TestSearchKeycloakUsersTruncatesTheMergedMatches(t *testing.T) {
	keycloak := &fakeKeycloak{users: []gocloak.User{
		keycloakUser("bob", "ab12cde.old", "bo34bbb"),
		keycloakUser("ada", "ada", "ab12cde"),
	}}

	results, err := keycloak.start(t).SearchKeycloakUsers(context.Background(), "ab12cde", 1)

	require.NoError(t, err)
	assert.Equal(t, []string{"ada"}, resultIDs(results))
	assert.True(t, results.Truncated)
}
