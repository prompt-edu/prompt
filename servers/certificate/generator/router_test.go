package generator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/certificate/config"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/prompt-edu/prompt/servers/certificate/participants"
	"github.com/prompt-edu/prompt/servers/certificate/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GeneratorRouterTestSuite struct {
	suite.Suite
	router          *gin.Engine
	suiteCtx        context.Context
	cleanup         func()
	mockCoreCleanup func()
	service         *GeneratorService
}

func (s *GeneratorRouterTestSuite) SetupSuite() {
	s.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(s.suiteCtx, "../database_dumps/certificate.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		s.T().Fatalf("Failed to set up test database: %v", err)
	}
	s.cleanup = cleanup

	_, mockCoreCleanup := testutils.SetupMockCoreService()
	s.mockCoreCleanup = mockCoreCleanup

	coreURL := "http://localhost:8080"
	if val, ok := os.LookupEnv("SERVER_CORE_HOST"); ok {
		coreURL = val
	}

	participantsService := participants.NewParticipantsService(*testDB.Queries, coreURL)
	s.service = NewGeneratorService(*testDB.Queries, config.NewConfigService(*testDB.Queries), participantsService, participantsService)

	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	api := s.router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, s.service, sdkTestUtils.MockPermissionMiddleware)
}

func (s *GeneratorRouterTestSuite) TearDownSuite() {
	if s.mockCoreCleanup != nil {
		s.mockCoreCleanup()
	}
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *GeneratorRouterTestSuite) TestDownloadStudentCertificate_InvalidCoursePhaseID() {
	url := "/api/course_phase/not-a-uuid/certificate/download/30000000-0000-0000-0000-000000000001"

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func (s *GeneratorRouterTestSuite) TestDownloadStudentCertificate_InvalidStudentID() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	url := fmt.Sprintf("/api/course_phase/%s/certificate/download/not-a-uuid", coursePhaseID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func (s *GeneratorRouterTestSuite) TestGetCertificateStatus_InvalidCoursePhaseID() {
	url := "/api/course_phase/not-a-uuid/certificate/status"

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func (s *GeneratorRouterTestSuite) TestGetTemplateConfig_WithTemplate() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")

	config, err := s.service.getTemplateConfig(s.newGinContext(), coursePhaseID)
	assert.NoError(s.T(), err)
	assert.True(s.T(), config.TemplateContent.Valid)
	assert.NotEmpty(s.T(), config.TemplateContent.String)
}

func (s *GeneratorRouterTestSuite) TestGetTemplateConfig_WithoutTemplate() {
	// Phase 2 exists but has NULL template
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000002")

	_, err := s.service.getTemplateConfig(s.newGinContext(), coursePhaseID)
	assert.ErrorIs(s.T(), err, errTemplateNotConfigured)
}

func (s *GeneratorRouterTestSuite) TestGetTemplateConfig_NonExistent() {
	// A missing config row is "not configured", not a genuine query error
	nonExistentID := uuid.New()

	_, err := s.service.getTemplateConfig(s.newGinContext(), nonExistentID)
	assert.ErrorIs(s.T(), err, errTemplateNotConfigured)
}

func (s *GeneratorRouterTestSuite) TestRecordCertificateDownload_Upsert() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	studentID := uuid.New()

	// First download
	d1, err := s.service.queries.RecordCertificateDownload(s.suiteCtx, db.RecordCertificateDownloadParams{
		StudentID:     studentID,
		CoursePhaseID: coursePhaseID,
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int32(1), d1.DownloadCount)

	// Second download increments
	d2, err := s.service.queries.RecordCertificateDownload(s.suiteCtx, db.RecordCertificateDownloadParams{
		StudentID:     studentID,
		CoursePhaseID: coursePhaseID,
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int32(2), d2.DownloadCount)
}

func (s *GeneratorRouterTestSuite) TestGetCertificateDownload_ExistingRecord() {
	studentID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")

	download, err := s.service.queries.GetCertificateDownload(s.suiteCtx, db.GetCertificateDownloadParams{
		StudentID:     studentID,
		CoursePhaseID: coursePhaseID,
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int32(3), download.DownloadCount)
}

func (s *GeneratorRouterTestSuite) TestCertificateStatusEndpoint_NoTemplate() {
	// Use a phase that has no template configured
	newPhaseID := uuid.New()

	url := fmt.Sprintf("/api/course_phase/%s/certificate/status", newPhaseID)
	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	// Without a configured template, the endpoint returns 200 with available=false
	assert.Equal(s.T(), http.StatusOK, resp.Code)

	var result map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), false, result["available"])
	assert.Equal(s.T(), false, result["hasDownloaded"])
}

// staffRouter authenticates every request as a course lecturer. The default suite
// router installs a no-op permission middleware, which leaves no token user, and
// the status handler needs one for every branch past "template not configured".
func (s *GeneratorRouterTestSuite) staffRouter() *gin.Engine {
	router := gin.New()
	api := router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, s.service, func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			keycloakTokenVerifier.SetTokenUser(c, keycloakTokenVerifier.TokenUser{
				Roles: map[string]bool{keycloakTokenVerifier.CourseLecturer: true},
			})
			c.Next()
		}
	})
	return router
}

func (s *GeneratorRouterTestSuite) certificateStatus(router *gin.Engine, coursePhaseID uuid.UUID) map[string]interface{} {
	url := fmt.Sprintf("/api/course_phase/%s/certificate/status", coursePhaseID)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(s.T(), http.StatusOK, resp.Code)

	var result map[string]interface{}
	assert.NoError(s.T(), json.Unmarshal(resp.Body.Bytes(), &result))
	return result
}

func (s *GeneratorRouterTestSuite) TestCertificateStatus_CarriesStudentPageText() {
	configService := config.NewConfigService(s.service.queries)
	text := "<p>Add it to your LinkedIn profile.</p>"

	// Without a template: the instructor's text is exactly what explains the wait.
	noTemplatePhase := uuid.New()
	_, err := configService.UpdateStudentPageText(s.suiteCtx, noTemplatePhase, &text, "Lecturer")
	assert.NoError(s.T(), err)

	status := s.certificateStatus(s.router, noTemplatePhase)
	assert.Equal(s.T(), false, status["available"])
	assert.Equal(s.T(), text, status["studentPageText"])

	// With a template configured, it is carried on the available status too.
	withTemplatePhase := uuid.New()
	_, err = configService.UpdateCoursePhaseConfig(s.suiteCtx, withTemplatePhase, "= Certificate", "Lecturer")
	assert.NoError(s.T(), err)
	_, err = configService.UpdateStudentPageText(s.suiteCtx, withTemplatePhase, &text, "Lecturer")
	assert.NoError(s.T(), err)

	staff := s.staffRouter()
	status = s.certificateStatus(staff, withTemplatePhase)
	assert.Equal(s.T(), true, status["available"])
	assert.Equal(s.T(), text, status["studentPageText"])

	// Cleared again, the key disappears rather than turning into an empty string.
	_, err = configService.UpdateStudentPageText(s.suiteCtx, withTemplatePhase, nil, "Lecturer")
	assert.NoError(s.T(), err)

	status = s.certificateStatus(staff, withTemplatePhase)
	_, present := status["studentPageText"]
	assert.False(s.T(), present)
}

func (s *GeneratorRouterTestSuite) TestCertificateStatus_OmitsStudentPageTextWhenUnset() {
	status := s.certificateStatus(s.router, uuid.New())
	_, present := status["studentPageText"]
	assert.False(s.T(), present)
}

// newGinContext creates a minimal gin.Context for unit-testing service functions.
func (s *GeneratorRouterTestSuite) newGinContext() *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}

// Verify JSON error response helper
func assertJSONError(t *testing.T, body []byte) {
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(t, err)
	_, hasError := result["error"]
	assert.True(t, hasError, "Response should contain an error field")
}

// --- Preview endpoint tests ---

func (s *GeneratorRouterTestSuite) TestPreviewCertificate_InvalidCoursePhaseID() {
	url := "/api/course_phase/not-a-uuid/certificate/preview"

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
	assertJSONError(s.T(), resp.Body.Bytes())
}

func (s *GeneratorRouterTestSuite) TestPreviewCertificate_NoTemplate() {
	// Phase 2 has NULL template
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000002")
	url := fmt.Sprintf("/api/course_phase/%s/certificate/preview", coursePhaseID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusNotFound, resp.Code)
	assertJSONError(s.T(), resp.Body.Bytes())
}

func (s *GeneratorRouterTestSuite) TestPreviewCertificate_NonExistentPhase() {
	nonExistentID := uuid.New()
	url := fmt.Sprintf("/api/course_phase/%s/certificate/preview", nonExistentID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusNotFound, resp.Code)
}

func (s *GeneratorRouterTestSuite) TestDownloadOwnCertificate_NoTemplate() {
	// Phase 2 has NULL template
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000002")
	url := fmt.Sprintf("/api/course_phase/%s/certificate/download", coursePhaseID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusNotFound, resp.Code)
	assertJSONError(s.T(), resp.Body.Bytes())
}

func (s *GeneratorRouterTestSuite) TestDownloadStudentCertificate_NoTemplate() {
	// Phase 2 has NULL template
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000002")
	studentID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	url := fmt.Sprintf("/api/course_phase/%s/certificate/download/%s", coursePhaseID, studentID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusNotFound, resp.Code)
	assertJSONError(s.T(), resp.Body.Bytes())
}

func (s *GeneratorRouterTestSuite) TestPreviewCertificate_Success() {
	if _, err := exec.LookPath("typst"); err != nil {
		s.T().Skip("typst binary not available, skipping preview test")
	}

	// Phase 1 has a valid template
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	url := fmt.Sprintf("/api/course_phase/%s/certificate/preview", coursePhaseID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusOK, resp.Code)
	assert.Equal(s.T(), "application/pdf", resp.Header().Get("Content-Type"))
	assert.Contains(s.T(), resp.Header().Get("Content-Disposition"), "certificate_preview.pdf")
	// PDF files start with %PDF
	assert.GreaterOrEqual(s.T(), len(resp.Body.Bytes()), 4, "PDF output should be at least 4 bytes")
	assert.Equal(s.T(), "%PDF", string(resp.Body.Bytes()[:4]))
}

func (s *GeneratorRouterTestSuite) TestPreviewCertificate_CompilationError() {
	if _, err := exec.LookPath("typst"); err != nil {
		s.T().Skip("typst binary not available, skipping preview test")
	}

	// Phase 3 has an invalid template (references nonexistent.json)
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000003")
	url := fmt.Sprintf("/api/course_phase/%s/certificate/preview", coursePhaseID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusUnprocessableEntity, resp.Code)

	var result map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Template compilation failed", result["error"])
	compilerOutput, ok := result["compilerOutput"].(string)
	assert.True(s.T(), ok, "Response should contain compilerOutput")
	assert.Contains(s.T(), compilerOutput, "nonexistent.json")
}

// --- Unit tests for helper functions ---

func (s *GeneratorRouterTestSuite) TestWriteDataFiles() {
	tempDir, err := os.MkdirTemp("", "test-write-data-*")
	assert.NoError(s.T(), err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	certData := CertificateData{
		StudentName: "Alice Smith",
		CourseName:  "Test Course",
		TeamName:    "Alpha",
		Date:        "January 1, 2026",
	}

	err = writeDataFiles(tempDir, certData)
	assert.NoError(s.T(), err)

	// Both files should exist with identical content
	for _, name := range []string{"data.json", "vars.json"} {
		content, err := os.ReadFile(filepath.Join(tempDir, name))
		assert.NoError(s.T(), err)

		var parsed CertificateData
		err = json.Unmarshal(content, &parsed)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), certData, parsed)
	}
}

func (s *GeneratorRouterTestSuite) TestTypstCompilationError_Interface() {
	typstErr := &TypstCompilationError{Output: "error: unknown variable"}
	assert.Contains(s.T(), typstErr.Error(), "unknown variable")

	// Should be detectable with errors.As
	var target *TypstCompilationError
	assert.True(s.T(), errors.As(typstErr, &target))
	assert.Equal(s.T(), "error: unknown variable", target.Output)
}

func (s *GeneratorRouterTestSuite) TestCompileTypst_InvalidTemplate() {
	if _, err := exec.LookPath("typst"); err != nil {
		s.T().Skip("typst binary not available, skipping compilation test")
	}

	tempDir, err := os.MkdirTemp("", "test-compile-*")
	assert.NoError(s.T(), err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Write an invalid template
	templatePath := filepath.Join(tempDir, "bad.typ")
	err = os.WriteFile(templatePath, []byte(`#let x = json("missing.json")`), 0644)
	assert.NoError(s.T(), err)

	_, err = compileTypst(s.suiteCtx, tempDir, templatePath)
	assert.Error(s.T(), err)

	var typstErr *TypstCompilationError
	assert.True(s.T(), errors.As(err, &typstErr))
	assert.Contains(s.T(), typstErr.Output, "missing.json")
}

func (s *GeneratorRouterTestSuite) TestCompileTypst_ValidTemplate() {
	if _, err := exec.LookPath("typst"); err != nil {
		s.T().Skip("typst binary not available, skipping compilation test")
	}

	tempDir, err := os.MkdirTemp("", "test-compile-*")
	assert.NoError(s.T(), err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Write data files
	certData := CertificateData{
		StudentName: "Bob",
		CourseName:  "Testing",
		TeamName:    "Beta",
		Date:        "Today",
	}
	err = writeDataFiles(tempDir, certData)
	assert.NoError(s.T(), err)

	// Write a valid template
	templatePath := filepath.Join(tempDir, "good.typ")
	err = os.WriteFile(templatePath, []byte(`#let d = json("data.json")
#d.studentName - #d.courseName`), 0644)
	assert.NoError(s.T(), err)

	pdfData, err := compileTypst(s.suiteCtx, tempDir, templatePath)
	assert.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(pdfData), 4, "PDF output should be at least 4 bytes")
	assert.Equal(s.T(), "%PDF", string(pdfData[:4]))
}

func TestGeneratorRouterTestSuite(t *testing.T) {
	suite.Run(t, new(GeneratorRouterTestSuite))
}
