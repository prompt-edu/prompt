package permissionValidation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	seededCourseID              = "5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e11"
	seededCoursePhaseID         = "5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e13"
	seededCourseParticipationID = "5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e15"
	seededCourseIdentifier      = "ios24245-iPraktikum"
)

type MiddlewareTestSuite struct {
	suite.Suite
	cleanup func()
	service *ValidationService
}

type idCheck struct {
	name     string
	check    PermissionCheck
	seededID string
}

func (suite *MiddlewareTestSuite) SetupSuite() {
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(context.Background(), "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/permission_validation_test.sql")
	require.NoError(suite.T(), err)
	suite.cleanup = cleanup
	suite.service = NewValidationService(*testDB.Queries)
}

func (suite *MiddlewareTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *MiddlewareTestSuite) idChecks() []idCheck {
	return []idCheck{
		{name: "course", check: suite.service.CheckCoursePermission, seededID: seededCourseID},
		{name: "course phase", check: suite.service.CheckCoursePhasePermission, seededID: seededCoursePhaseID},
		{name: "course participation", check: suite.service.CheckCourseParticipationPermission, seededID: seededCourseParticipationID},
	}
}

func serveByID(check PermissionCheck, auth gin.HandlerFunc, id string, allowedRoles ...string) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/:id", auth, CheckAccessControlByID(check, "id", allowedRoles...), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"courseTokenIdentifier": c.GetString("courseTokenIdentifier")})
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/"+id, nil))
	return w
}

func serveAs(check PermissionCheck, roles []string, id string) *httptest.ResponseRecorder {
	return serveByID(check, sdkTestUtils.MockAuthMiddleware(roles), id, PromptAdmin, CourseLecturer)
}

func (suite *MiddlewareTestSuite) assertSingleJSONBody(w *httptest.ResponseRecorder, label string) {
	assert.True(suite.T(), json.Valid(w.Body.Bytes()), "%s returned %q", label, w.Body.String())
}

func (suite *MiddlewareTestSuite) TestUnknownIDIsNotFoundForPromptAdminOnly() {
	for _, tc := range suite.idChecks() {
		for role, expected := range map[string]int{
			PromptAdmin:                          http.StatusNotFound,
			seededCourseIdentifier + "-Lecturer": http.StatusForbidden,
			PromptLecturer:                       http.StatusForbidden,
		} {
			label := tc.name + " as " + role
			w := serveAs(tc.check, []string{role}, uuid.NewString())

			assert.Equal(suite.T(), expected, w.Code, label)
			suite.assertSingleJSONBody(w, label)
			assert.NotContains(suite.T(), w.Body.String(), "no rows", label)
		}
	}
}

func (suite *MiddlewareTestSuite) TestUnknownIDIsForbiddenForPromptAdminOnRoutesThatExcludeIt() {
	for _, tc := range suite.idChecks() {
		w := serveByID(tc.check, sdkTestUtils.MockAuthMiddleware([]string{PromptAdmin}), uuid.NewString(), CourseLecturer)

		assert.Equal(suite.T(), http.StatusForbidden, w.Code, tc.name)
		suite.assertSingleJSONBody(w, tc.name)
	}
}

func (suite *MiddlewareTestSuite) TestUserWithoutCourseRoleIsForbidden() {
	for _, tc := range suite.idChecks() {
		w := serveAs(tc.check, []string{"ws2425-other-Lecturer", seededCourseIdentifier + "-Student"}, tc.seededID)

		assert.Equal(suite.T(), http.StatusForbidden, w.Code, tc.name)
		suite.assertSingleJSONBody(w, tc.name)
	}
}

func (suite *MiddlewareTestSuite) TestMatchingRoleSetsCourseTokenIdentifier() {
	for _, tc := range suite.idChecks() {
		for _, role := range []string{PromptAdmin, seededCourseIdentifier + "-Lecturer"} {
			label := tc.name + " as " + role
			w := serveAs(tc.check, []string{role}, tc.seededID)

			require.Equal(suite.T(), http.StatusOK, w.Code, label)
			assert.JSONEq(suite.T(), `{"courseTokenIdentifier":"`+seededCourseIdentifier+`"}`, w.Body.String(), label)
		}
	}
}

func (suite *MiddlewareTestSuite) TestMissingUserRolesIsAnInternalServerError() {
	for _, tc := range suite.idChecks() {
		w := serveByID(tc.check, func(c *gin.Context) { c.Next() }, tc.seededID, PromptAdmin, CourseLecturer)

		assert.Equal(suite.T(), http.StatusInternalServerError, w.Code, tc.name)
		suite.assertSingleJSONBody(w, tc.name)
		assert.NotContains(suite.T(), w.Body.String(), "user roles", tc.name)
	}
}

func (suite *MiddlewareTestSuite) TestCheckErrorDoesNotLeakItsText() {
	failing := func(*gin.Context, uuid.UUID, ...string) (bool, error) {
		return false, errors.New("dial tcp 10.0.0.5:5432: connection refused")
	}

	w := serveAs(failing, []string{PromptAdmin}, uuid.NewString())

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	suite.assertSingleJSONBody(w, "failing check")
	assert.NotContains(suite.T(), w.Body.String(), "connection refused")
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}
