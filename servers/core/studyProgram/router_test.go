package studyProgram

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/studyProgram/studyProgramDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	suite.Suite
	ctx             context.Context
	cleanup         func()
	adminRouter     *gin.Engine
	lecturerRouter  *gin.Engine
	anonymousRouter *gin.Engine
}

func (suite *RouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/study_program_test.sql")
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	service := NewStudyProgramService(*testDB.Queries, testDB.Conn)

	suite.adminRouter = setupRouter(service, func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware([]string{permissionValidation.PromptAdmin})
	})
	suite.lecturerRouter = setupRouter(service, func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware([]string{permissionValidation.PromptLecturer})
	})
	suite.anonymousRouter = setupRouter(service, func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	})
}

func (suite *RouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func setupRouter(service *StudyProgramService, authMiddleware func() gin.HandlerFunc) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	setupStudyProgramRouter(api, service, authMiddleware, permissionValidation.CheckAccessControlByRole)
	return router
}

func serve(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func (suite *RouterTestSuite) TestListIsPublic() {
	w := serve(suite.anonymousRouter, http.MethodGet, "/api/study-programs", nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var studyPrograms []studyProgramDTO.StudyProgram
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &studyPrograms))
	assert.NotEmpty(suite.T(), studyPrograms)
}

func (suite *RouterTestSuite) TestWritesRequireAuthentication() {
	w := serve(suite.anonymousRouter, http.MethodPost, "/api/study-programs", studyProgramDTO.CreateStudyProgram{Name: "Physics"})
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	w = serve(suite.anonymousRouter, http.MethodGet, "/api/study-programs/student-counts", nil)
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *RouterTestSuite) TestWritesRequirePromptAdmin() {
	w := serve(suite.lecturerRouter, http.MethodPost, "/api/study-programs", studyProgramDTO.CreateStudyProgram{Name: "Physics"})
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	w = serve(suite.lecturerRouter, http.MethodPut, "/api/study-programs/"+computerScienceID.String(), studyProgramDTO.UpdateStudyProgram{Name: "Physics"})
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	w = serve(suite.lecturerRouter, http.MethodDelete, "/api/study-programs/"+computerScienceID.String(), nil)
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	w = serve(suite.lecturerRouter, http.MethodGet, "/api/study-programs/student-counts", nil)
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *RouterTestSuite) TestCreateStudyProgram() {
	w := serve(suite.adminRouter, http.MethodPost, "/api/study-programs", studyProgramDTO.CreateStudyProgram{
		Name:      " Mathematics ",
		ShortName: "MA",
	})

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	var created studyProgramDTO.StudyProgram
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(suite.T(), "Mathematics", created.Name)
	assert.Equal(suite.T(), "MA", created.ShortName.String)
}

func (suite *RouterTestSuite) TestCreateInvalidStudyProgram() {
	for _, input := range []studyProgramDTO.CreateStudyProgram{
		{Name: "  "},
		{Name: "Other"},
		{Name: " oTHer "},
		{Name: "Unknown"},
		{Name: "Computer Engineering", ShortName: "Other"},
		{Name: "Computer Engineering", ShortName: "unknown"},
	} {
		w := serve(suite.adminRouter, http.MethodPost, "/api/study-programs", input)
		assert.Equal(suite.T(), http.StatusBadRequest, w.Code, "name %q", input.Name)
	}
}

func (suite *RouterTestSuite) TestCreateDuplicateStudyProgram() {
	w := serve(suite.adminRouter, http.MethodPost, "/api/study-programs", studyProgramDTO.CreateStudyProgram{Name: "games engineering"})

	assert.Equal(suite.T(), http.StatusConflict, w.Code)
}

func (suite *RouterTestSuite) TestCreateStudyProgramWithTakenLabel() {
	w := serve(suite.adminRouter, http.MethodPost, "/api/study-programs", studyProgramDTO.CreateStudyProgram{
		Name:      "Physics",
		ShortName: "PH",
	})
	require.Equal(suite.T(), http.StatusCreated, w.Code)

	w = serve(suite.adminRouter, http.MethodPost, "/api/study-programs", studyProgramDTO.CreateStudyProgram{
		Name:      "Engineering Physics",
		ShortName: "ph",
	})

	assert.Equal(suite.T(), http.StatusConflict, w.Code)
	assert.Contains(suite.T(), w.Body.String(), ErrDuplicateStudyProgramLabel.Error())
}

func (suite *RouterTestSuite) TestUpdateStudyProgram() {
	w := serve(suite.adminRouter, http.MethodPut, "/api/study-programs/"+informationSystemsID.String(), studyProgramDTO.UpdateStudyProgram{
		Name:      "Information Systems",
		ShortName: "WI",
	})

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var updated studyProgramDTO.StudyProgram
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(suite.T(), "WI", updated.ShortName.String)
}

func (suite *RouterTestSuite) TestUpdateInvalidStudyProgram() {
	w := serve(suite.adminRouter, http.MethodPut, "/api/study-programs/"+informationSystemsID.String(), studyProgramDTO.UpdateStudyProgram{
		Name: "other",
	})

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *RouterTestSuite) TestDeleteStudyProgram() {
	w := serve(suite.adminRouter, http.MethodPost, "/api/study-programs", studyProgramDTO.CreateStudyProgram{Name: "Chemistry"})
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	var created studyProgramDTO.StudyProgram
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &created))

	w = serve(suite.adminRouter, http.MethodDelete, "/api/study-programs/"+created.ID.String(), nil)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
}

func (suite *RouterTestSuite) TestUpdateWithInvalidID() {
	w := serve(suite.adminRouter, http.MethodPut, "/api/study-programs/not-a-uuid", studyProgramDTO.UpdateStudyProgram{Name: "Physics"})

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *RouterTestSuite) TestUpdateUnknownStudyProgram() {
	w := serve(suite.adminRouter, http.MethodPut, "/api/study-programs/"+uuid.NewString(), studyProgramDTO.UpdateStudyProgram{Name: "Physics"})

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *RouterTestSuite) TestDeleteUnknownStudyProgram() {
	w := serve(suite.adminRouter, http.MethodDelete, "/api/study-programs/"+uuid.NewString(), nil)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *RouterTestSuite) TestGetStudentCounts() {
	w := serve(suite.adminRouter, http.MethodGet, "/api/study-programs/student-counts", nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var counts []studyProgramDTO.StudyProgramStudentCount
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &counts))
	assert.NotEmpty(suite.T(), counts)
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
