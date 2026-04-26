package privacy

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	authService "github.com/prompt-edu/prompt/servers/core/auth/service"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	"github.com/prompt-edu/prompt/servers/core/privacy/service"
	"github.com/prompt-edu/prompt/servers/core/storage"
	"github.com/prompt-edu/prompt/servers/core/storage/privacyexport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================================
// Test records – mirrors of export_test.sql seed data
// ============================================================

// User IDs (keycloak subject IDs, NOT student table PKs)
var (
	aliceUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	bobUserID   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	carolUserID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	daveUserID  = uuid.MustParse("44444444-4444-4444-4444-444444444444")
)

// Student IDs (FK into the student table)
var (
	aliceStudentID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	bobStudentID   = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	carolStudentID = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	daveStudentID  = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
)

// export4 – Dave's rate-limited export. Timestamps are relative (NOW() in SQL)
// and intentionally left as zero values since we never assert on them.
var export4 = privacyDTO.PrivacyExport{
	ID:        uuid.MustParse("e4444444-4444-4444-4444-444444444444"),
	UserID:    daveUserID,
	StudentID: &daveStudentID,
	Status:    db.ExportStatusComplete,
}

// export1 – all docs successful (complete / no_data), owner: Alice
var export1 = privacyDTO.PrivacyExport{
	ID:          uuid.MustParse("e1111111-1111-1111-1111-111111111111"),
	UserID:      aliceUserID,
	StudentID:   &aliceStudentID,
	Status:      db.ExportStatusComplete,
	DateCreated: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
	ValidUntil:  time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	Documents: []privacyDTO.PrivacyExportDocument{
		{
			ID:          uuid.MustParse("d1111111-1111-1111-1111-100000000001"),
			DateCreated: time.Date(2026, 3, 1, 10, 0, 1, 0, time.UTC),
			SourceName:  "Core",
			Status:      db.ExportStatusComplete,
			FileSize:    int64Ptr(2048),
		},
		{
			ID:          uuid.MustParse("d1111111-1111-1111-1111-100000000002"),
			DateCreated: time.Date(2026, 3, 1, 10, 0, 2, 0, time.UTC),
			SourceName:  "MicroserviceA",
			Status:      db.ExportStatusNoData,
		},
	},
}

// export2 – partially failed (complete + no_data + failed docs), owner: Bob
var export2 = privacyDTO.PrivacyExport{
	ID:          uuid.MustParse("e2222222-2222-2222-2222-222222222222"),
	UserID:      bobUserID,
	StudentID:   &bobStudentID,
	Status:      db.ExportStatusFailed,
	DateCreated: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC),
	ValidUntil:  time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	Documents: []privacyDTO.PrivacyExportDocument{
		{
			ID:          uuid.MustParse("d2222222-2222-2222-2222-200000000001"),
			DateCreated: time.Date(2026, 3, 5, 10, 0, 1, 0, time.UTC),
			SourceName:  "Core",
			Status:      db.ExportStatusComplete,
			FileSize:    int64Ptr(1024),
		},
		{
			ID:          uuid.MustParse("d2222222-2222-2222-2222-200000000002"),
			DateCreated: time.Date(2026, 3, 5, 10, 0, 2, 0, time.UTC),
			SourceName:  "MicroserviceA",
			Status:      db.ExportStatusNoData,
		},
		{
			ID:          uuid.MustParse("d2222222-2222-2222-2222-200000000003"),
			DateCreated: time.Date(2026, 3, 5, 10, 0, 3, 0, time.UTC),
			SourceName:  "MicroserviceB",
			Status:      db.ExportStatusFailed,
		},
	},
}

// export3 – archived (deletion routine ran, S3 files gone), owner: Carol
// valid_until is in the past — that is what triggers the deletion routine
var export3 = privacyDTO.PrivacyExport{
	ID:          uuid.MustParse("e3333333-3333-3333-3333-333333333333"),
	UserID:      carolUserID,
	StudentID:   &carolStudentID,
	Status:      db.ExportStatusArchived,
	DateCreated: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
	ValidUntil:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	Documents: []privacyDTO.PrivacyExportDocument{
		{
			ID:          uuid.MustParse("d3333333-3333-3333-3333-300000000001"),
			DateCreated: time.Date(2026, 2, 1, 10, 0, 1, 0, time.UTC),
			SourceName:  "Core",
			Status:      db.ExportStatusArchived,
			FileSize:    int64Ptr(2048),
		},
		{
			ID:          uuid.MustParse("d3333333-3333-3333-3333-300000000002"),
			DateCreated: time.Date(2026, 2, 1, 10, 0, 2, 0, time.UTC),
			SourceName:  "MicroserviceA",
			Status:      db.ExportStatusArchived,
			FileSize:    int64Ptr(512),
		},
	},
}

// AdminPrivacyExport variants – Docs are returned as a minimal list ordered by source_name.
// None of the seed docs have downloaded_at set, so Downloaded is false for all.
var (
	adminExport1 = privacyDTO.AdminPrivacyExport{
		ID:          export1.ID,
		Status:      export1.Status,
		DateCreated: export1.DateCreated,
		ValidUntil:  export1.ValidUntil,
		Docs: []privacyDTO.AdminExportDoc{
			{SourceName: "Core", Status: db.ExportStatusComplete, Downloaded: false},
			{SourceName: "MicroserviceA", Status: db.ExportStatusNoData, Downloaded: false},
		},
	}
	adminExport2 = privacyDTO.AdminPrivacyExport{
		ID:          export2.ID,
		Status:      export2.Status,
		DateCreated: export2.DateCreated,
		ValidUntil:  export2.ValidUntil,
		Docs: []privacyDTO.AdminExportDoc{
			{SourceName: "Core", Status: db.ExportStatusComplete, Downloaded: false},
			{SourceName: "MicroserviceA", Status: db.ExportStatusNoData, Downloaded: false},
			{SourceName: "MicroserviceB", Status: db.ExportStatusFailed, Downloaded: false},
		},
	}
	adminExport3 = privacyDTO.AdminPrivacyExport{
		ID:          export3.ID,
		Status:      export3.Status,
		DateCreated: export3.DateCreated,
		ValidUntil:  export3.ValidUntil,
		Docs: []privacyDTO.AdminExportDoc{
			{SourceName: "Core", Status: db.ExportStatusArchived, Downloaded: false},
			{SourceName: "MicroserviceA", Status: db.ExportStatusArchived, Downloaded: false},
		},
	}
	adminExport4 = privacyDTO.AdminPrivacyExport{
		ID:     export4.ID,
		Status: export4.Status,
		Docs: []privacyDTO.AdminExportDoc{
			{SourceName: "Core", Status: db.ExportStatusComplete, Downloaded: false},
		},
	}
)

func int64Ptr(v int64) *int64 { return &v }

// ============================================================
// Suite setup
// ============================================================

type RouterTestSuite struct {
	suite.Suite
	router         *gin.Engine
	ctx            context.Context
	cleanup        func()
}

func (suite *RouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	// Set up the test database
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/export_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup

	service.InitPrivacyService(*testDB.Queries, testDB.Conn)
	authService.InitAuthService(*testDB.Queries, testDB.Conn)
	privacyexport.InitWithAdapter(&storage.MockStorageAdapter{})

	suite.router = setupRouter()
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	authMiddleware := func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware([]string{"PROMPT_Admin"})
	}
	permissionMiddleware := sdkTestUtils.MockPermissionMiddleware
	setupPrivacyRouter(api, authMiddleware, permissionMiddleware)
	return router
}

// routerForUser returns a router that authenticates requests as the given user ID.
// Required for endpoints that call GetUserUUIDFromContext (all non-admin routes).
func routerForUser(userID uuid.UUID) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	authMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("userRoles", map[string]bool{"PROMPT_Admin": true})
			c.Set("userID", userID.String())
			c.Next()
		}
	}
	setupPrivacyRouter(api, authMiddleware, sdkTestUtils.MockPermissionMiddleware)
	return router
}

func (suite *RouterTestSuite) TestRouterGetAllExportsAdmin() {
	req, _ := http.NewRequest("GET", "/api/privacy/admin/data-exports", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var exports []privacyDTO.AdminPrivacyExport
	err := json.Unmarshal(w.Body.Bytes(), &exports)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), exports, 4)

	// DB orders by date_created DESC:
	//   [0] adminExport4 – Dave, created NOW()-5d (always most recent)
	//   [1] adminExport2 – Bob,   Mar 5
	//   [2] adminExport1 – Alice, Mar 1
	//   [3] adminExport3 – Carol, Feb 1
	expected := []privacyDTO.AdminPrivacyExport{adminExport4, adminExport2, adminExport1, adminExport3}
	for i, exp := range expected {
		assert.Equal(suite.T(), exp.ID, exports[i].ID)
		assert.Equal(suite.T(), exp.Status, exports[i].Status)
		assert.Equal(suite.T(), exp.Docs, exports[i].Docs)
	}
}

func (suite *RouterTestSuite) TestRouterGetLatestExport_Exists() {
	// Alice has a valid export (valid_until: 2030)
	router := routerForUser(aliceUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var body map[string]json.RawMessage
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `"exists"`, string(body["status"]))

	var exp privacyDTO.PrivacyExport
	err = json.Unmarshal(body["export"], &exp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), export1.ID, exp.ID)
	assert.Equal(suite.T(), export1.Status, exp.Status)
}

func (suite *RouterTestSuite) TestRouterGetLatestExport_RateLimited() {
	// Dave's export expired yesterday but was created 5 days ago — still within the 30-day rate limit
	router := routerForUser(daveUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var body map[string]json.RawMessage
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `"rate_limited"`, string(body["status"]))

	// retry_after must be a timestamp in the future
	var retryAfter time.Time
	err = json.Unmarshal(body["retry_after"], &retryAfter)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), retryAfter.After(time.Now()))
}

func (suite *RouterTestSuite) TestRouterGetLatestExport_NoExport() {
	// Carol's export is archived and past the rate limit window → 204
	router := routerForUser(carolUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
}

func (suite *RouterTestSuite) TestRouterGetExport_Owner() {
	router := routerForUser(aliceUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export/"+export1.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var result privacyDTO.PrivacyExport
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), export1.ID, result.ID)
	assert.Equal(suite.T(), export1.Status, result.Status)
	assert.Len(suite.T(), result.Documents, len(export1.Documents))
	for i, doc := range export1.Documents {
		assert.Equal(suite.T(), doc.ID, result.Documents[i].ID)
		assert.Equal(suite.T(), doc.Status, result.Documents[i].Status)
	}
}

func (suite *RouterTestSuite) TestRouterGetExport_WrongUser() {
	// Bob tries to fetch Alice's export
	router := routerForUser(bobUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export/"+export1.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *RouterTestSuite) TestRouterGetExport_Archived() {
	// Carol fetches her own export but valid_until is in the past
	router := routerForUser(carolUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export/"+export3.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *RouterTestSuite) TestRouterPostDataExport_ValidExportAlreadyExists() {
	// Alice already has a valid export (valid_until: 2030)
	router := routerForUser(aliceUserID)
	req, _ := http.NewRequest("POST", "/api/privacy/data-export", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
}

func (suite *RouterTestSuite) TestRouterPostDataExport_RateLimited() {
	// Dave's export expired but he is still within the 30-day rate limit window
	router := routerForUser(daveUserID)
	req, _ := http.NewRequest("POST", "/api/privacy/data-export", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
}

func (suite *RouterTestSuite) TestRouterGetDownloadURL_Owner() {
	docID := export1.Documents[0].ID // 'complete' doc belonging to Alice
	router := routerForUser(aliceUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export/"+export1.ID.String()+"/docs/"+docID.String()+"/download-url", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), body["downloadUrl"])
}

func (suite *RouterTestSuite) TestRouterGetDownloadURL_WrongUser() {
	// Bob tries to get a download URL for Alice's export doc
	docID := export1.Documents[0].ID
	router := routerForUser(bobUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export/"+export1.ID.String()+"/docs/"+docID.String()+"/download-url", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *RouterTestSuite) TestRouterGetDownloadURL_DocNotInExport() {
	// Alice requests a URL using a doc ID that belongs to a different export (export2)
	foreignDocID := export2.Documents[0].ID
	router := routerForUser(aliceUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export/"+export1.ID.String()+"/docs/"+foreignDocID.String()+"/download-url", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *RouterTestSuite) TestRouterGetDownloadURL_ArchivedExport() {
	// Carol tries to download from her own archived (expired) export
	docID := export3.Documents[0].ID
	router := routerForUser(carolUserID)
	req, _ := http.NewRequest("GET", "/api/privacy/data-export/"+export3.ID.String()+"/docs/"+docID.String()+"/download-url", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
