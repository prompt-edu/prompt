package copy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/course/copy/courseCopyDTO"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The orgs ase and aet from copy_course_test.sql, both created on 2026-01-01. ase owns ASE Template
// and ASE Course; Master Test is ungrouped.
var (
	aseOrgID       = uuid.MustParse("0a1e0000-0000-4000-8000-000000000001")
	aetOrgID       = uuid.MustParse("0a1e0000-0000-4000-8000-000000000002")
	aseTemplateID  = "0a1e0000-0000-4000-8000-000000000011"
	aseCourseID    = "0a1e0000-0000-4000-8000-000000000012"
	ungroupedID    = "c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd"
	issuedLater    = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	copiedCourseNo = 0
)

func (suite *CourseCopyRouterTestSuite) orgRouter(roles ...string) *gin.Engine {
	router := gin.Default()
	validation := permissionValidation.NewValidationService(suite.courseCopyService.queries)
	setupCourseCopyRouter(router.Group("/api"), suite.courseCopyService, func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(keycloakTokenVerifier.CtxTokenIssuedAt, issuedLater)
			sdkTestUtils.MockAuthMiddleware(roles)(c)
		}
	}, permissionValidation.CheckAccessControlByRole, checkAccessControlByIDWrapper(validation.CheckCoursePermission))
	return router
}

func (suite *CourseCopyRouterTestSuite) copyAs(roles []string, sourceID string, orgID *uuid.UUID) *httptest.ResponseRecorder {
	copiedCourseNo++
	body, _ := json.Marshal(courseCopyDTO.CopyCourseRequest{
		Name:        fmt.Sprintf("OrgCopy%d", copiedCourseNo),
		SemesterTag: pgtype.Text{String: "ws2627", Valid: true},
		StartDate:   pgtype.Date{Valid: true, Time: time.Now()},
		EndDate:     pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
		OrgID:       orgID,
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/courses/"+sourceID+"/copy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.orgRouter(roles...).ServeHTTP(w, req)
	return w
}

func (suite *CourseCopyRouterTestSuite) TestCopyChecksTheOrg() {
	for _, tt := range []struct {
		name     string
		roles    []string
		sourceID string
		orgID    *uuid.UUID
		want     int
	}{
		{"org member copies the org's template into the org", []string{"org-ase-Member"}, aseTemplateID, &aseOrgID, http.StatusCreated},
		{"org member copies the org's template without an org", []string{"org-ase-Member"}, aseTemplateID, nil, http.StatusForbidden},
		{"org member copies the org's template into another org", []string{"org-ase-Member"}, aseTemplateID, &aetOrgID, http.StatusForbidden},
		{"org member copies a course of the org", []string{"org-ase-Member"}, aseCourseID, &aseOrgID, http.StatusForbidden},
		{"member of another org copies the template", []string{"org-aet-Member"}, aseTemplateID, &aetOrgID, http.StatusForbidden},
		{"org admin copies the org's course into the org", []string{"org-ase-Admin"}, aseCourseID, &aseOrgID, http.StatusCreated},
		{"org admin copies the org's course without an org", []string{"org-ase-Admin"}, aseCourseID, nil, http.StatusForbidden},
		{"course lecturer copies their course without an org", []string{"ss25-Master Test-Lecturer"}, ungroupedID, nil, http.StatusCreated},
		{"course lecturer copies their course into an org they are not in", []string{"ss25-Master Test-Lecturer"}, ungroupedID, &aseOrgID, http.StatusForbidden},
	} {
		w := suite.copyAs(tt.roles, tt.sourceID, tt.orgID)

		require.Equal(suite.T(), tt.want, w.Code, "%s: %s", tt.name, w.Body.String())
		if tt.want == http.StatusCreated {
			var created courseDTO.Course
			require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &created), tt.name)
			assert.Equal(suite.T(), tt.orgID, created.OrgID, "the copy's org comes from the request: %s", tt.name)
		}
	}
}

func (suite *CourseCopyRouterTestSuite) TestOrgMemberChecksATemplateIsCopyable() {
	for sourceID, want := range map[string]int{aseTemplateID: http.StatusOK, aseCourseID: http.StatusForbidden} {
		req, _ := http.NewRequest(http.MethodGet, "/api/courses/"+sourceID+"/copyable", nil)
		w := httptest.NewRecorder()
		suite.orgRouter("org-ase-Member").ServeHTTP(w, req)

		assert.Equal(suite.T(), want, w.Code, sourceID)
	}
}
