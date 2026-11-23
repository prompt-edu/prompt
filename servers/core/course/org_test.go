package course

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt-sdk/audit"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The orgs ase and aet from copy_course_test.sql, both created on 2026-01-01; ase owns ASE Template
// and ASE Course.
var (
	aseOrgID      = uuid.MustParse("0a1e0000-0000-4000-8000-000000000001")
	aetOrgID      = uuid.MustParse("0a1e0000-0000-4000-8000-000000000002")
	aseTemplateID = uuid.MustParse("0a1e0000-0000-4000-8000-000000000011")
	aseCourseID   = uuid.MustParse("0a1e0000-0000-4000-8000-000000000012")

	issuedAfterOrgs  = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	issuedBeforeOrgs = time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
)

type auditCapture chan audit.Event

func (a auditCapture) Record(_ context.Context, e audit.Event) error {
	a <- e
	return nil
}

// orgRouter serves the course routes with the real permission checks for a caller with the given
// roles and token issue time.
func (suite *CourseRouterTestSuite) orgRouter(issuedAt time.Time, roles []string, middleware ...gin.HandlerFunc) *gin.Engine {
	router := gin.Default()
	validation := permissionValidation.NewValidationService(suite.courseService.queries)
	setupCourseRouter(router.Group("/api", middleware...), suite.courseService, func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(keycloakTokenVerifier.CtxTokenIssuedAt, issuedAt)
			sdkTestUtils.MockAuthMiddleware(roles)(c)
		}
	}, permissionValidation.CheckAccessControlByRole, checkAccessControlByIDWrapper(validation.CheckCoursePermission))
	return router
}

func serveJSON(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func newOrgCourse(name string, orgID *uuid.UUID) courseDTO.CreateCourse {
	return courseDTO.CreateCourse{
		Name:        name,
		StartDate:   pgtype.Date{Valid: true, Time: time.Now()},
		EndDate:     pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
		SemesterTag: pgtype.Text{String: "ws2627", Valid: true},
		CourseType:  "practical course",
		Ects:        pgtype.Int4{Int32: 10, Valid: true},
		OrgID:       orgID,
	}
}

func (suite *CourseRouterTestSuite) TestCreateCourseChecksTheOrg() {
	unknownOrgID := uuid.New()
	for i, tt := range []struct {
		name     string
		roles    []string
		issuedAt time.Time
		orgID    *uuid.UUID
		want     int
	}{
		{"org member in their org", []string{"org-ase-Member"}, issuedAfterOrgs, &aseOrgID, http.StatusCreated},
		{"org admin in their org", []string{"org-ase-Admin"}, issuedAfterOrgs, &aseOrgID, http.StatusCreated},
		{"PROMPT_Lecturer without an org", []string{permissionValidation.PromptLecturer}, issuedAfterOrgs, nil, http.StatusCreated},
		{"PROMPT_Admin in any org", []string{permissionValidation.PromptAdmin}, issuedAfterOrgs, &aetOrgID, http.StatusCreated},
		{"org member without an org", []string{"org-ase-Member"}, issuedAfterOrgs, nil, http.StatusForbidden},
		{"org member in another org", []string{"org-ase-Member"}, issuedAfterOrgs, &aetOrgID, http.StatusForbidden},
		{"org member with a token from before the org existed", []string{"org-ase-Member"}, issuedBeforeOrgs, &aseOrgID, http.StatusForbidden},
		{"PROMPT_Lecturer in an org they are not in", []string{permissionValidation.PromptLecturer}, issuedAfterOrgs, &aseOrgID, http.StatusForbidden},
		{"course lecturer", []string{"ss25-Master Test-Lecturer"}, issuedAfterOrgs, &aseOrgID, http.StatusForbidden},
		{"custom-group role that looks like an org role", []string{"org-ase-cg-Member"}, issuedAfterOrgs, &aseOrgID, http.StatusForbidden},
		{"PROMPT_Admin in an unknown org", []string{permissionValidation.PromptAdmin}, issuedAfterOrgs, &unknownOrgID, http.StatusBadRequest},
	} {
		w := serveJSON(suite.orgRouter(tt.issuedAt, tt.roles), http.MethodPost, "/api/courses/", newOrgCourse(fmt.Sprintf("OrgCourse%d", i), tt.orgID))

		require.Equal(suite.T(), tt.want, w.Code, "%s: %s", tt.name, w.Body.String())
		if tt.want == http.StatusCreated {
			var created courseDTO.Course
			require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &created), tt.name)
			assert.Equal(suite.T(), tt.orgID, created.OrgID, tt.name)
		}
	}
}

func (suite *CourseRouterTestSuite) TestCheckCourseNameIsOpenToOrgRoles() {
	for roles, want := range map[string]int{
		"org-ase-Member":            http.StatusOK,
		"org-ase-Admin":             http.StatusOK,
		"ss25-Master Test-Lecturer": http.StatusForbidden,
	} {
		w := serveJSON(suite.orgRouter(issuedAfterOrgs, []string{roles}), http.MethodGet, "/api/courses/check-name?name=Free&semesterTag=ws2627", nil)

		assert.Equal(suite.T(), want, w.Code, roles)
	}
}

func (suite *CourseRouterTestSuite) courseOrg(courseID uuid.UUID) *uuid.UUID {
	course, err := suite.courseService.queries.GetCourse(suite.ctx, courseID)
	require.NoError(suite.T(), err)
	return courseDTO.OrgIDFromDBModel(course.OrgID)
}

func (suite *CourseRouterTestSuite) TestUpdateCourseOrg() {
	defer func() {
		_, err := suite.courseService.UpdateCourseOrg(suite.ctx, aseCourseID, &aseOrgID)
		require.NoError(suite.T(), err)
	}()
	events := make(auditCapture, 1)
	router := suite.orgRouter(issuedAfterOrgs, []string{permissionValidation.PromptAdmin}, audit.Middleware(events, audit.WithActorExtractor(func(*gin.Context) (audit.Actor, bool) {
		return audit.Actor{ID: "admin"}, true
	})))
	path := "/api/courses/" + aseCourseID.String() + "/org"

	w := serveJSON(router, http.MethodPut, path, map[string]any{"orgID": aetOrgID})
	require.Equal(suite.T(), http.StatusNoContent, w.Code, w.Body.String())
	assert.Equal(suite.T(), &aetOrgID, suite.courseOrg(aseCourseID))
	event := suite.expectAuditEvent(events)
	assert.Equal(suite.T(), map[string]any{"orgID": aetOrgID.String(), "previousOrgID": aseOrgID.String()}, event.Metadata)

	w = serveJSON(router, http.MethodPut, path, map[string]any{"orgID": nil})
	require.Equal(suite.T(), http.StatusNoContent, w.Code, w.Body.String())
	assert.Nil(suite.T(), suite.courseOrg(aseCourseID))
	event = suite.expectAuditEvent(events)
	assert.Equal(suite.T(), map[string]any{"previousOrgID": aetOrgID.String()}, event.Metadata)

	for _, tt := range []struct {
		name string
		path string
		body any
		want int
	}{
		{"without the orgID key", path, map[string]any{}, http.StatusBadRequest},
		{"into an unknown org", path, map[string]any{"orgID": uuid.New()}, http.StatusBadRequest},
		{"of an unknown course", "/api/courses/" + uuid.NewString() + "/org", map[string]any{"orgID": nil}, http.StatusNotFound},
	} {
		w := serveJSON(router, http.MethodPut, tt.path, tt.body)
		assert.Equal(suite.T(), tt.want, w.Code, tt.name)
	}
}

func (suite *CourseRouterTestSuite) expectAuditEvent(events auditCapture) audit.Event {
	select {
	case event := <-events:
		assert.Equal(suite.T(), "Assigned a course to an org", event.Action)
		assert.Equal(suite.T(), "course", event.EntityType)
		assert.Equal(suite.T(), aseCourseID.String(), event.EntityID)
		assert.Equal(suite.T(), "ASE Course", event.EntityName)
		return event
	case <-time.After(5 * time.Second):
		suite.T().Fatal("no audit event")
		return audit.Event{}
	}
}

func (suite *CourseRouterTestSuite) TestOnlyPromptAdminsAssignCoursesToOrgs() {
	for _, roles := range [][]string{{"org-ase-Admin"}, {"org-aet-Admin"}, {permissionValidation.PromptLecturer}, {"ss25-ASE Course-Lecturer"}} {
		w := serveJSON(suite.orgRouter(issuedAfterOrgs, roles), http.MethodPut, "/api/courses/"+aseCourseID.String()+"/org", map[string]any{"orgID": aetOrgID})

		assert.Equal(suite.T(), http.StatusForbidden, w.Code, roles)
	}
	assert.Equal(suite.T(), &aseOrgID, suite.courseOrg(aseCourseID))
}

func (suite *CourseRouterTestSuite) TestOrgAdminManagesTheOrgsCourses() {
	path := "/api/courses/" + aseCourseID.String()
	for roles, want := range map[string]int{
		"org-ase-Admin":  http.StatusOK,
		"org-ase-Member": http.StatusForbidden,
		"org-aet-Admin":  http.StatusForbidden,
	} {
		w := serveJSON(suite.orgRouter(issuedAfterOrgs, []string{roles}), http.MethodGet, path, nil)

		assert.Equal(suite.T(), want, w.Code, roles)
	}
}

// listed returns the courses the path lists for the caller, by ID with their org. Other tests of the
// suite add courses to the orgs, so the assertions only rely on the seeded ones.
func (suite *CourseRouterTestSuite) listed(issuedAt time.Time, roles []string, path string) map[uuid.UUID]*uuid.UUID {
	w := serveJSON(suite.orgRouter(issuedAt, roles), http.MethodGet, path, nil)
	require.Equal(suite.T(), http.StatusOK, w.Code, "%v: %s", roles, w.Body.String())
	var courses []courseDTO.Course
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &courses))
	orgByCourse := map[uuid.UUID]*uuid.UUID{}
	for _, course := range courses {
		orgByCourse[course.ID] = course.OrgID
	}
	return orgByCourse
}

func (suite *CourseRouterTestSuite) TestOrgAdminListsAllCoursesOfTheOrg() {
	listed := suite.listed(issuedAfterOrgs, []string{"org-ase-Admin"}, "/api/courses/")
	assert.Contains(suite.T(), listed, aseTemplateID)
	assert.Contains(suite.T(), listed, aseCourseID)
	for courseID, orgID := range listed {
		assert.Equal(suite.T(), &aseOrgID, orgID, "%s is not a course of the org", courseID)
	}

	for name, tt := range map[string]struct {
		roles    []string
		issuedAt time.Time
	}{
		"org member":                         {[]string{"org-ase-Member"}, issuedAfterOrgs},
		"admin of another org":               {[]string{"org-aet-Admin"}, issuedAfterOrgs},
		"org admin with a token from before": {[]string{"org-ase-Admin"}, issuedBeforeOrgs},
	} {
		listed := suite.listed(tt.issuedAt, tt.roles, "/api/courses/")
		assert.NotContains(suite.T(), listed, aseTemplateID, name)
		assert.NotContains(suite.T(), listed, aseCourseID, name)
	}
}

func (suite *CourseRouterTestSuite) TestOrgRolesListTheOrgsTemplates() {
	for _, roles := range [][]string{{"org-ase-Member"}, {"org-ase-Admin"}} {
		listed := suite.listed(issuedAfterOrgs, roles, "/api/courses/template")
		assert.Contains(suite.T(), listed, aseTemplateID, roles)
		assert.NotContains(suite.T(), listed, aseCourseID, roles)
	}
	assert.NotContains(suite.T(), suite.listed(issuedAfterOrgs, []string{"org-aet-Member"}, "/api/courses/template"), aseTemplateID)
	assert.NotContains(suite.T(), suite.listed(issuedBeforeOrgs, []string{"org-ase-Member"}, "/api/courses/template"), aseTemplateID)
}
