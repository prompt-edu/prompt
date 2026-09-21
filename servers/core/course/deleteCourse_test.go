package course

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingPhaseProvider struct{}

func (failingPhaseProvider) GetCoursePhaseByID(context.Context, uuid.UUID) (coursePhaseDTO.CoursePhase, error) {
	return coursePhaseDTO.CoursePhase{}, nil
}

func (failingPhaseProvider) CheckCoursePhasesBelongToCourse(context.Context, uuid.UUID, []uuid.UUID) (bool, error) {
	return true, nil
}

func (failingPhaseProvider) DeleteModuleDataForCourse(context.Context, string, uuid.UUID) error {
	return errors.New("the team allocation module could not delete its data")
}

// A module that cannot drop its data must stop the course deletion before anything else is
// touched, so the whole operation stays retryable. The queries and the pool are unset on
// purpose: reaching either of them would mean the short circuit did not work.
func TestDeleteCourseStopsWhenAModuleFails(t *testing.T) {
	keycloakDeleted := false
	service := NewCourseService(db.Queries{}, nil, failingPhaseProvider{},
		func(context.Context, string, string, string) error { return nil },
		func(context.Context, uuid.UUID) error {
			keycloakDeleted = true
			return nil
		},
	)

	err := service.DeleteCourse(context.Background(), "Bearer token", uuid.New())

	require.Error(t, err)
	assert.False(t, keycloakDeleted, "the Keycloak groups must survive a failed module deletion")
}
