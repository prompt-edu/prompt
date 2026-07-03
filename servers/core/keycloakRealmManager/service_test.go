package keycloakRealmManager

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

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
	err := AddUserToCourseGroup(context.Background(), uuid.New(), "Admin", "user-1", "caller-1")
	if !errors.Is(err, ErrInvalidGroupName) {
		t.Errorf("expected ErrInvalidGroupName, got %v", err)
	}
}

func TestRemoveUserFromCourseGroup_RejectsBadGroupName(t *testing.T) {
	err := RemoveUserFromCourseGroup(context.Background(), uuid.New(), "Admin", "user-1", "caller-1")
	if !errors.Is(err, ErrInvalidGroupName) {
		t.Errorf("expected ErrInvalidGroupName, got %v", err)
	}
}

func TestRemoveUserFromCourseGroup_RejectsSelfRemoval(t *testing.T) {
	const userID = "abc-123"
	err := RemoveUserFromCourseGroup(context.Background(), uuid.New(), permissionValidation.CourseLecturer, userID, userID)
	if !errors.Is(err, ErrSelfRemoval) {
		t.Errorf("expected ErrSelfRemoval, got %v", err)
	}
}

func TestSearchKeycloakUsers_RejectsShortQuery(t *testing.T) {
	cases := []string{"", " ", "a", "  x  "}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			_, err := SearchKeycloakUsers(context.Background(), q, 20)
			if !errors.Is(err, ErrInvalidQuery) {
				t.Errorf("expected ErrInvalidQuery for %q, got %v", q, err)
			}
		})
	}
}
