package permissionValidation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The org ase owns the course Agile (with a phase and a participation) and the template
// AgileTemplate; its child org ase-lab owns the course Lab. Both orgs were created on 2026-01-01.
const (
	orgCourseID              = "5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e23"
	orgCoursePhaseID         = "5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e24"
	orgCourseParticipationID = "5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e25"
	orgTemplateID            = "5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e26"
	childOrgCourseID         = "5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e27"
	orgCourseIdentifier      = "ws2526-Agile"

	aseAdmin  = "org-ase-Admin"
	aseMember = "org-ase-Member"
)

var (
	issuedAfterOrgs  = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	issuedBeforeOrgs = time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
)

type orgCheckResult struct {
	Status       int
	OrgAdmin     bool `json:"courseOrgAdmin"`
	AccessViaOrg bool `json:"courseAccessViaOrg"`
}

func serveWithToken(check PermissionCheck, issuedAt time.Time, roles []string, id string, allowedRoles ...string) orgCheckResult {
	router := gin.New()
	router.GET("/:id", func(c *gin.Context) {
		c.Set(keycloakTokenVerifier.CtxTokenIssuedAt, issuedAt)
		sdkTestUtils.MockAuthMiddleware(roles)(c)
	}, CheckAccessControlByID(check, "id", allowedRoles...), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{CtxCourseOrgAdmin: c.GetBool(CtxCourseOrgAdmin), CtxCourseAccessViaOrg: c.GetBool(CtxCourseAccessViaOrg)})
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/"+id, nil))

	result := orgCheckResult{Status: w.Code}
	if w.Code == http.StatusOK {
		_ = json.Unmarshal(w.Body.Bytes(), &result)
	}
	return result
}

func (suite *MiddlewareTestSuite) orgCourseChecks() []idCheck {
	return []idCheck{
		{name: "course", check: suite.service.CheckCoursePermission, seededID: orgCourseID},
		{name: "course phase", check: suite.service.CheckCoursePhasePermission, seededID: orgCoursePhaseID},
		{name: "course participation", check: suite.service.CheckCourseParticipationPermission, seededID: orgCourseParticipationID},
	}
}

func (suite *MiddlewareTestSuite) TestOrgAdminIsLecturerEquivalentOnTheOrgsCourses() {
	for _, tc := range suite.orgCourseChecks() {
		result := serveWithToken(tc.check, issuedAfterOrgs, []string{aseAdmin}, tc.seededID, PromptAdmin, CourseLecturer)

		require.Equal(suite.T(), http.StatusOK, result.Status, tc.name)
		assert.True(suite.T(), result.OrgAdmin, tc.name)
		assert.True(suite.T(), result.AccessViaOrg, tc.name)
	}
}

func (suite *MiddlewareTestSuite) TestOrgAdminPassesOnlyWhereLecturersDo() {
	for _, tc := range suite.orgCourseChecks() {
		result := serveWithToken(tc.check, issuedAfterOrgs, []string{aseAdmin}, tc.seededID, PromptAdmin, CourseEditor, CourseStudent)

		assert.Equal(suite.T(), http.StatusForbidden, result.Status, tc.name)
	}
}

func (suite *MiddlewareTestSuite) TestOrgRolesThatGrantNothingOnACourse() {
	for _, tt := range []struct {
		name     string
		roles    []string
		issuedAt time.Time
		courseID string
	}{
		{"org member on the org's course", []string{aseMember}, issuedAfterOrgs, orgCourseID},
		{"org admin on an ungrouped course", []string{aseAdmin}, issuedAfterOrgs, seededCourseID},
		{"parent org admin on the child org's course", []string{aseAdmin}, issuedAfterOrgs, childOrgCourseID},
		{"child org admin on the parent org's course", []string{"org-ase-lab-Admin"}, issuedAfterOrgs, orgCourseID},
		{"org admin with a token from before the org existed", []string{aseAdmin}, issuedBeforeOrgs, orgCourseID},
		{"custom-group role of a course named like the org", []string{"org-ase-cg-Admin"}, issuedAfterOrgs, orgCourseID},
	} {
		result := serveWithToken(suite.service.CheckCoursePermission, tt.issuedAt, tt.roles, tt.courseID, PromptAdmin, CourseLecturer, CourseEditor)

		assert.Equal(suite.T(), http.StatusForbidden, result.Status, tt.name)
	}
}

func (suite *MiddlewareTestSuite) TestCourseRoleOnAnOrgCourseIsNotAccessViaOrg() {
	result := serveWithToken(suite.service.CheckCoursePermission, issuedAfterOrgs, []string{orgCourseIdentifier + "-Lecturer", aseAdmin}, orgCourseID, PromptAdmin, CourseLecturer)

	require.Equal(suite.T(), http.StatusOK, result.Status)
	assert.True(suite.T(), result.OrgAdmin)
	assert.False(suite.T(), result.AccessViaOrg)
}

func (suite *MiddlewareTestSuite) TestOrgMemberMayUseTheOrgsTemplatesOnly() {
	for _, tt := range []struct {
		name     string
		roles    []string
		courseID string
		want     int
	}{
		{"org member on the org's template", []string{aseMember}, orgTemplateID, http.StatusOK},
		{"org member on the org's course", []string{aseMember}, orgCourseID, http.StatusForbidden},
		{"member of the child org on the parent org's template", []string{"org-ase-lab-Member"}, orgTemplateID, http.StatusForbidden},
		{"org admin is no member", []string{aseAdmin}, orgTemplateID, http.StatusForbidden},
	} {
		result := serveWithToken(suite.service.CheckCoursePermission, issuedAfterOrgs, tt.roles, tt.courseID, PromptAdmin, OrgMember)

		assert.Equal(suite.T(), tt.want, result.Status, tt.name)
		if tt.want == http.StatusOK {
			assert.True(suite.T(), result.AccessViaOrg, tt.name)
		}
	}

	result := serveWithToken(suite.service.CheckCoursePermission, issuedBeforeOrgs, []string{aseMember}, orgTemplateID, PromptAdmin, OrgMember)
	assert.Equal(suite.T(), http.StatusForbidden, result.Status, "org member with a token from before the org existed")
}
