package coursePhase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const testAuthHeader = "Bearer module-deletion-token"

// fakeModule stands in for a course phase module: it answers the SDK's /info probe and the
// phase deletion endpoint, and records what core sent.
type fakeModule struct {
	server       *httptest.Server
	info         http.HandlerFunc
	deleteStatus int

	mu          sync.Mutex
	infoCalls   int
	deletedIDs  []string
	authHeaders []string
}

func newFakeModule(info http.HandlerFunc, deleteStatus int) *fakeModule {
	m := &fakeModule{info: info, deleteStatus: deleteStatus}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/info":
			m.infoCalls++
			m.info(w, r)
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/course_phase/"):
			m.deletedIDs = append(m.deletedIDs, strings.TrimPrefix(r.URL.Path, "/course_phase/"))
			m.authHeaders = append(m.authHeaders, r.Header.Get("Authorization"))
			w.WriteHeader(m.deleteStatus)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return m
}

func supportsDeletion(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`{"serviceName":"fake","healthy":true,"capabilities":{"phase.deletion":true}}`))
}

func withoutDeletion(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`{"serviceName":"fake","healthy":true,"capabilities":{"phase.copy":true}}`))
}

// spaFallback mimics a reverse proxy with no route for the module: the client's single page
// document is served under status 200.
func spaFallback(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte("<!doctype html><html><body>PROMPT</body></html>"))
}

func statusOnly(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }
}

type ModuleDeletionTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	conn    *pgxpool.Pool
	queries *db.Queries
	service *CoursePhaseService
}

func (suite *ModuleDeletionTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/course_phase_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err, "failed to set up test database")

	suite.cleanup = cleanup
	suite.conn = testDB.Conn
	suite.queries = testDB.Queries
	suite.service = NewCoursePhaseService(*testDB.Queries, testDB.Conn, resolution.NewResolutionService("localhost:8080"))
}

func (suite *ModuleDeletionTestSuite) TearDownSuite() {
	suite.cleanup()
}

// newPhaseType inserts a phase type with the given base URL. Every test gets its own so the
// tests never depend on each other's rows.
func (suite *ModuleDeletionTestSuite) newPhaseType(baseURL string) uuid.UUID {
	id := uuid.New()
	_, err := suite.conn.Exec(suite.ctx,
		`INSERT INTO course_phase_type (id, name, base_url) VALUES ($1, $2, $3)`,
		id, "module-deletion-"+id.String(), baseURL)
	require.NoError(suite.T(), err)
	return id
}

func (suite *ModuleDeletionTestSuite) newPhase(phaseTypeID, courseID uuid.UUID) uuid.UUID {
	id := uuid.New()
	_, err := suite.conn.Exec(suite.ctx,
		`INSERT INTO course_phase (id, course_id, name, is_initial_phase, course_phase_type_id)
		 VALUES ($1, $2, 'Module Deletion Phase', false, $3)`,
		id, courseID, phaseTypeID)
	require.NoError(suite.T(), err)
	return id
}

func (suite *ModuleDeletionTestSuite) phaseExists(id uuid.UUID) bool {
	var exists bool
	err := suite.conn.QueryRow(suite.ctx, `SELECT EXISTS (SELECT 1 FROM course_phase WHERE id = $1)`, id).Scan(&exists)
	require.NoError(suite.T(), err)
	return exists
}

func (suite *ModuleDeletionTestSuite) TestDeletesModuleDataThenPhase() {
	module := newFakeModule(supportsDeletion, http.StatusOK)
	defer module.server.Close()
	phaseID := suite.newPhase(suite.newPhaseType(module.server.URL), uuid.New())

	err := suite.service.DeleteCoursePhase(suite.ctx, testAuthHeader, phaseID)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), suite.phaseExists(phaseID), "the phase row should be gone")
	assert.Equal(suite.T(), 1, module.infoCalls)
	assert.Equal(suite.T(), []string{phaseID.String()}, module.deletedIDs)
	assert.Equal(suite.T(), []string{testAuthHeader}, module.authHeaders, "the caller's token must be forwarded")
}

func (suite *ModuleDeletionTestSuite) TestSkipsModuleWithoutTheCapability() {
	module := newFakeModule(withoutDeletion, http.StatusOK)
	defer module.server.Close()
	phaseID := suite.newPhase(suite.newPhaseType(module.server.URL), uuid.New())

	err := suite.service.DeleteCoursePhase(suite.ctx, testAuthHeader, phaseID)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), suite.phaseExists(phaseID))
	assert.Empty(suite.T(), module.deletedIDs, "a module without the capability must not be called")
}

func (suite *ModuleDeletionTestSuite) TestSkipsCoreImplementedPhase() {
	module := newFakeModule(supportsDeletion, http.StatusOK)
	defer module.server.Close()
	phaseID := suite.newPhase(suite.newPhaseType("core"), uuid.New())

	err := suite.service.DeleteCoursePhase(suite.ctx, testAuthHeader, phaseID)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), suite.phaseExists(phaseID))
	assert.Zero(suite.T(), module.infoCalls, "a core implemented phase type has no module to ask")
}

func (suite *ModuleDeletionTestSuite) TestProbesOnceForPhasesSharingAModule() {
	module := newFakeModule(supportsDeletion, http.StatusOK)
	defer module.server.Close()
	phaseTypeID := suite.newPhaseType(module.server.URL)
	courseID := uuid.New()
	first := suite.newPhase(phaseTypeID, courseID)
	second := suite.newPhase(phaseTypeID, courseID)

	err := suite.service.DeleteModuleDataForCourse(suite.ctx, testAuthHeader, courseID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, module.infoCalls, "one capability probe per module, not per phase")
	assert.ElementsMatch(suite.T(), []string{first.String(), second.String()}, module.deletedIDs)
}

func (suite *ModuleDeletionTestSuite) TestDeletesEveryModuleOfACourse() {
	first := newFakeModule(supportsDeletion, http.StatusOK)
	defer first.server.Close()
	second := newFakeModule(supportsDeletion, http.StatusOK)
	defer second.server.Close()

	courseID := uuid.New()
	firstPhase := suite.newPhase(suite.newPhaseType(first.server.URL), courseID)
	secondPhase := suite.newPhase(suite.newPhaseType(second.server.URL), courseID)

	err := suite.service.DeleteModuleDataForCourse(suite.ctx, testAuthHeader, courseID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{firstPhase.String()}, first.deletedIDs)
	assert.Equal(suite.T(), []string{secondPhase.String()}, second.deletedIDs)
}

func (suite *ModuleDeletionTestSuite) TestResolvesTheCoreHostPlaceholder() {
	module := newFakeModule(supportsDeletion, http.StatusOK)
	defer module.server.Close()

	service := NewCoursePhaseService(*suite.queries, suite.conn, resolution.NewResolutionService(module.server.URL))
	phaseID := suite.newPhase(suite.newPhaseType("{CORE_HOST}"), uuid.New())

	err := service.DeleteCoursePhase(suite.ctx, testAuthHeader, phaseID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{phaseID.String()}, module.deletedIDs)
}

func (suite *ModuleDeletionTestSuite) TestKeepsThePhaseWhenTheDeletionFails() {
	// Everything the module can answer other than 200 leaves core unsure whether the data is
	// gone, including the 405 an unproxied path yields and a redirect to somewhere else.
	for name, status := range map[string]int{
		"not found":          http.StatusNotFound,
		"method not allowed": http.StatusMethodNotAllowed,
		"server error":       http.StatusInternalServerError,
		"redirect":           http.StatusFound,
	} {
		suite.Run(name, func() {
			module := newFakeModule(supportsDeletion, status)
			defer module.server.Close()
			phaseID := suite.newPhase(suite.newPhaseType(module.server.URL), uuid.New())

			err := suite.service.DeleteCoursePhase(suite.ctx, testAuthHeader, phaseID)

			assert.Error(suite.T(), err)
			assert.True(suite.T(), suite.phaseExists(phaseID), "the phase must survive so the deletion can be retried")
		})
	}
}

func (suite *ModuleDeletionTestSuite) TestKeepsThePhaseWhenTheCapabilityIsUnknown() {
	for name, info := range map[string]http.HandlerFunc{
		"no info endpoint":   statusOnly(http.StatusNotFound),
		"info server error":  statusOnly(http.StatusInternalServerError),
		"proxy spa fallback": spaFallback,
	} {
		suite.Run(name, func() {
			module := newFakeModule(info, http.StatusOK)
			defer module.server.Close()
			phaseID := suite.newPhase(suite.newPhaseType(module.server.URL), uuid.New())

			err := suite.service.DeleteCoursePhase(suite.ctx, testAuthHeader, phaseID)

			assert.Error(suite.T(), err)
			assert.True(suite.T(), suite.phaseExists(phaseID))
			assert.Empty(suite.T(), module.deletedIDs)
		})
	}
}

func (suite *ModuleDeletionTestSuite) TestKeepsThePhaseWhenTheModuleIsUnreachable() {
	module := newFakeModule(supportsDeletion, http.StatusOK)
	unreachableURL := module.server.URL
	module.server.Close()

	phaseID := suite.newPhase(suite.newPhaseType(unreachableURL), uuid.New())

	err := suite.service.DeleteCoursePhase(suite.ctx, testAuthHeader, phaseID)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), suite.phaseExists(phaseID))
}

func (suite *ModuleDeletionTestSuite) TestKeepsThePhaseWhenTheBaseURLIsUnusable() {
	phaseID := suite.newPhase(suite.newPhaseType("not a url"), uuid.New())

	err := suite.service.DeleteCoursePhase(suite.ctx, testAuthHeader, phaseID)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), suite.phaseExists(phaseID))
}

func TestModuleDeletionTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleDeletionTestSuite))
}
