package course

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubPhaseProvider reports the given phases as cleaned up by their modules, or fails with err.
type stubPhaseProvider struct {
	cleaned []uuid.UUID
	err     error
}

func (stubPhaseProvider) GetCoursePhaseByID(context.Context, uuid.UUID) (coursePhaseDTO.CoursePhase, error) {
	return coursePhaseDTO.CoursePhase{}, nil
}

func (stubPhaseProvider) CheckCoursePhasesBelongToCourse(context.Context, uuid.UUID, []uuid.UUID) (bool, error) {
	return true, nil
}

func (p stubPhaseProvider) DeleteModuleDataForCourse(context.Context, string, uuid.UUID) ([]uuid.UUID, error) {
	return p.cleaned, p.err
}

// A module that cannot drop its data must stop the course deletion before anything else is
// touched, so the whole operation stays retryable. The queries and the pool are unset on
// purpose: reaching either of them would mean the short circuit did not work.
func TestDeleteCourseStopsWhenAModuleFails(t *testing.T) {
	keycloakDeleted := false
	provider := stubPhaseProvider{err: errors.New("the team allocation module could not delete its data")}
	service := NewCourseService(db.Queries{}, nil, provider,
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

// The phases of the seeded course in course_test.sql.
var (
	seededCourseID = uuid.MustParse("3f42d322-e5bf-4faa-b576-51f2cab14c2e")
	seededPhaseIDs = []uuid.UUID{
		uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411b"),
		uuid.MustParse("92bb0532-39e5-453d-bc50-fa61ea0128b2"),
		uuid.MustParse("500db7ed-2eb2-42d0-82b3-8750e12afa8a"),
	}
)

func TestDeleteCourseChecksForPhasesAddedDuringTheModuleCleanup(t *testing.T) {
	ctx := context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(ctx, "../database_dumps/course_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(t, err)
	defer cleanup()

	newService := func(cleaned []uuid.UUID, keycloakDeleted *bool) *CourseService {
		return NewCourseService(*testDB.Queries, testDB.Conn, stubPhaseProvider{cleaned: cleaned},
			func(context.Context, string, string, string) error { return nil },
			func(context.Context, uuid.UUID) error {
				*keycloakDeleted = true
				return nil
			},
		)
	}

	t.Run("a phase the modules were not asked about keeps the course", func(t *testing.T) {
		keycloakDeleted := false

		err := newService(seededPhaseIDs[:2], &keycloakDeleted).DeleteCourse(ctx, "Bearer token", seededCourseID)

		require.ErrorIs(t, err, ErrCourseChangedDuringDeletion)
		assert.False(t, keycloakDeleted, "the Keycloak groups must survive a refused deletion")
		_, err = testDB.Queries.GetCourse(ctx, seededCourseID)
		assert.NoError(t, err, "the course must survive so the deletion can be retried")
	})

	t.Run("every phase covered deletes the course", func(t *testing.T) {
		keycloakDeleted := false

		err := newService(seededPhaseIDs, &keycloakDeleted).DeleteCourse(ctx, "Bearer token", seededCourseID)

		require.NoError(t, err)
		assert.True(t, keycloakDeleted)
		_, err = testDB.Queries.GetCourse(ctx, seededCourseID)
		assert.Error(t, err, "the course must be gone")
	})
}
