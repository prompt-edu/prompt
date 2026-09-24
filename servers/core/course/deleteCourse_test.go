package course

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
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

func newDeleteCourseRouter(service *CourseService) *gin.Engine {
	router := gin.New()
	setupCourseRouter(router.Group("/api"), service, func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware([]string{"PROMPT_Admin"})
	}, sdkTestUtils.MockPermissionMiddleware, sdkTestUtils.MockPermissionMiddleware)
	return router
}

func deleteCourseRequest(router *gin.Engine, courseID uuid.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/courses/"+courseID.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

// A module that cannot drop its data must stop the course deletion before anything else is
// touched, so the whole operation stays retryable. The queries and the pool are unset on
// purpose: reaching either of them would mean the short circuit did not work.
func TestDeleteCourseStopsWhenAModuleFails(t *testing.T) {
	keycloakDeleted := false
	moduleErr := fmt.Errorf("%w: http://team-allocation:8083/api/course_phase/x answered 500: secret detail", coursePhase.ErrModuleDeletionFailed)
	service := NewCourseService(db.Queries{}, nil, stubPhaseProvider{err: moduleErr},
		func(context.Context, string, string, string) error { return nil },
		func(context.Context, uuid.UUID) error {
			keycloakDeleted = true
			return nil
		},
	)

	resp := deleteCourseRequest(newDeleteCourseRouter(service), uuid.New())

	assert.Equal(t, http.StatusBadGateway, resp.Code, "a module failure is an upstream failure")
	assert.NotContains(t, resp.Body.String(), "team-allocation", "the module URL must not reach the client")
	assert.NotContains(t, resp.Body.String(), "secret detail", "the module's answer must not reach the client")
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

	// The Keycloak deletion runs while the course row is locked, so it must see the deadline that
	// bounds the transaction.
	keycloakHadDeadline := false
	newService := func(cleaned []uuid.UUID, keycloakDeleted *bool) *CourseService {
		return NewCourseService(*testDB.Queries, testDB.Conn, stubPhaseProvider{cleaned: cleaned},
			func(context.Context, string, string, string) error { return nil },
			func(ctx context.Context, _ uuid.UUID) error {
				*keycloakDeleted = true
				_, keycloakHadDeadline = ctx.Deadline()
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
		assert.True(t, keycloakHadDeadline, "the transaction holding the course lock must be bounded")
		_, err = testDB.Queries.GetCourse(ctx, seededCourseID)
		assert.Error(t, err, "the course must be gone")
	})
}
