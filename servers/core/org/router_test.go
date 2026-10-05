package org

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
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// The router tests run the real role and org permission checks instead of
// MockPermissionMiddleware, so the permission matrix is tested as it ships.
type OrgRouterTestSuite struct {
	suite.Suite
	ctx               context.Context
	cleanup           func()
	service           *OrgService
	validationService *permissionValidation.ValidationService
	org               orgDTO.Org
	childOrg          orgDTO.Org
	otherOrg          orgDTO.Org
}

func (suite *OrgRouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := setupMigratedTestDB(suite.ctx)
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	noKeycloak := func(context.Context, string) error { return nil }
	suite.service = NewOrgService(*testDB.Queries, testDB.Conn, noKeycloak, noKeycloak)
	suite.validationService = permissionValidation.NewValidationService(*testDB.Queries)

	suite.org = suite.createOrg("ase", nil)
	suite.childOrg = suite.createOrg("ase-lab", &suite.org.ID)
	suite.otherOrg = suite.createOrg("aet", nil)
}

func (suite *OrgRouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *OrgRouterTestSuite) createOrg(slug string, parentOrgID *uuid.UUID) orgDTO.Org {
	created, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{ParentOrgID: parentOrgID, Name: "Org " + slug, Slug: slug})
	require.NoError(suite.T(), err)
	return created
}

func (suite *OrgRouterTestSuite) routerFor(roles ...string) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	setupOrgRouter(api, suite.service, func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware(roles)
	}, permissionValidation.CheckAccessControlByRole, checkAccessControlByIDWrapper(suite.validationService.CheckOrgPermission))
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

func orgPath(orgID uuid.UUID) string {
	return "/api/orgs/" + orgID.String()
}

type actor struct {
	name  string
	roles []string
}

var (
	promptAdmin = actor{name: "PROMPT_Admin", roles: []string{permissionValidation.PromptAdmin}}
	orgAdmin    = actor{name: "org admin", roles: []string{"org-ase-Admin"}}
	orgMember   = actor{name: "org member", roles: []string{"org-ase-Member"}}
	otherAdmin  = actor{name: "admin of another org", roles: []string{"org-aet-Admin"}}
	customGroup = actor{name: "custom group role ending in Admin", roles: []string{"org-ase-cg-Admin"}}
	lecturer    = actor{name: "PROMPT_Lecturer", roles: []string{permissionValidation.PromptLecturer}}
	noOrgRoles  = actor{name: "user without org roles", roles: []string{"ws25-ase-Lecturer"}}
	allActors   = []actor{promptAdmin, orgAdmin, orgMember, otherAdmin, customGroup, lecturer, noOrgRoles}
)

// expectStatus sends the request as each actor, building the path and body afresh for
// every request, and expects allowedStatus for the allowed actors and 403 for everyone
// else. Every response must carry exactly one JSON body.
func (suite *OrgRouterTestSuite) expectStatus(method string, path func() string, body func() any, allowedStatus int, allowed ...actor) {
	for _, a := range allActors {
		expected := http.StatusForbidden
		for _, allowedActor := range allowed {
			if allowedActor.name == a.name {
				expected = allowedStatus
			}
		}

		requestPath := path()
		w := serve(suite.routerFor(a.roles...), method, requestPath, body())

		assert.Equal(suite.T(), expected, w.Code, "%s %s as %s", method, requestPath, a.name)
		if w.Body.Len() > 0 {
			assert.True(suite.T(), json.Valid(w.Body.Bytes()), "%s %s as %s returned %q", method, requestPath, a.name, w.Body.String())
		}
	}
}

func noBody() any { return nil }

func (suite *OrgRouterTestSuite) TestListOrgsPermissions() {
	suite.expectStatus(http.MethodGet, func() string { return "/api/orgs/" }, noBody, http.StatusOK, promptAdmin)
}

func (suite *OrgRouterTestSuite) TestGetOrgPermissions() {
	suite.expectStatus(http.MethodGet, func() string { return orgPath(suite.org.ID) }, noBody, http.StatusOK, promptAdmin, orgAdmin, orgMember)
}

func (suite *OrgRouterTestSuite) TestParentOrgRolesGrantNothingOnChildOrgs() {
	suite.expectStatus(http.MethodGet, func() string { return orgPath(suite.childOrg.ID) }, noBody, http.StatusOK, promptAdmin)
	suite.expectStatus(http.MethodPut, func() string { return orgPath(suite.childOrg.ID) }, func() any {
		return orgDTO.UpdateOrg{Name: "Org ase-lab"}
	}, http.StatusOK, promptAdmin)
}

func (suite *OrgRouterTestSuite) TestUpdateOrgPermissions() {
	suite.expectStatus(http.MethodPut, func() string { return orgPath(suite.org.ID) }, func() any {
		return orgDTO.UpdateOrg{Name: "Applied Software Engineering"}
	}, http.StatusOK, promptAdmin, orgAdmin)
}

func (suite *OrgRouterTestSuite) TestCreateOrgPermissions() {
	suite.expectStatus(http.MethodPost, func() string { return "/api/orgs/" }, func() any {
		return orgDTO.CreateOrg{Name: "Created Org", Slug: "created-" + uuid.NewString()[:8]}
	}, http.StatusCreated, promptAdmin)
}

func (suite *OrgRouterTestSuite) TestUpdateOrgParentPermissions() {
	suite.expectStatus(http.MethodPut, func() string { return orgPath(suite.otherOrg.ID) + "/parent" }, func() any {
		return orgDTO.UpdateOrgParent{ParentOrgID: nil}
	}, http.StatusOK, promptAdmin)
}

func (suite *OrgRouterTestSuite) TestDeleteOrgPermissions() {
	suite.expectStatus(http.MethodDelete, func() string {
		return orgPath(suite.createOrg("disposable-"+uuid.NewString()[:8], nil).ID)
	}, noBody, http.StatusNoContent, promptAdmin)
}

func (suite *OrgRouterTestSuite) TestUnknownOrgIsNotFoundForAdminsOnly() {
	path := orgPath(uuid.New())

	assert.Equal(suite.T(), http.StatusNotFound, serve(suite.routerFor(promptAdmin.roles...), http.MethodGet, path, nil).Code)
	assert.Equal(suite.T(), http.StatusForbidden, serve(suite.routerFor(orgAdmin.roles...), http.MethodGet, path, nil).Code)
	assert.Equal(suite.T(), http.StatusNotFound, serve(suite.routerFor(promptAdmin.roles...), http.MethodDelete, path, nil).Code)
}

func (suite *OrgRouterTestSuite) TestInvalidOrgIDIsABadRequest() {
	w := serve(suite.routerFor(orgAdmin.roles...), http.MethodGet, "/api/orgs/not-a-uuid", nil)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *OrgRouterTestSuite) TestCreateOrgValidatesInput() {
	admin := suite.routerFor(promptAdmin.roles...)

	for name, input := range map[string]orgDTO.CreateOrg{
		"missing name":  {Slug: "valid-slug"},
		"invalid slug":  {Name: "Org", Slug: "Not A Slug"},
		"cg segment":    {Name: "Org", Slug: "ase-cg"},
		"invalid email": {Name: "Org", Slug: "valid-slug", ContactEmail: "office"},
		"invalid URL":   {Name: "Org", Slug: "valid-slug", Website: "example.com"},
	} {
		w := serve(admin, http.MethodPost, "/api/orgs/", input)
		assert.Equal(suite.T(), http.StatusBadRequest, w.Code, name)
	}
}

func (suite *OrgRouterTestSuite) TestCreateOrgWithDuplicateSlugIsAConflict() {
	w := serve(suite.routerFor(promptAdmin.roles...), http.MethodPost, "/api/orgs/", orgDTO.CreateOrg{Name: "Duplicate", Slug: "ase"})

	assert.Equal(suite.T(), http.StatusConflict, w.Code)
}

func (suite *OrgRouterTestSuite) TestCreateOrgUnderUnknownParentIsABadRequest() {
	unknown := uuid.New()

	w := serve(suite.routerFor(promptAdmin.roles...), http.MethodPost, "/api/orgs/", orgDTO.CreateOrg{ParentOrgID: &unknown, Name: "Orphan", Slug: "orphan"})

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *OrgRouterTestSuite) TestMovingAnOrgWithoutAParentKeyIsABadRequest() {
	for name, body := range map[string]any{
		"empty body":     map[string]any{},
		"misspelled key": map[string]any{"parentOrg": nil},
	} {
		w := serve(suite.routerFor(promptAdmin.roles...), http.MethodPut, orgPath(suite.childOrg.ID)+"/parent", body)
		assert.Equal(suite.T(), http.StatusBadRequest, w.Code, name)
	}
	assert.Equal(suite.T(), &suite.org.ID, suite.parentOf(suite.childOrg.ID), "the org stays under its parent")
}

func (suite *OrgRouterTestSuite) parentOf(orgID uuid.UUID) *uuid.UUID {
	org, err := suite.service.GetOrg(suite.ctx, orgID)
	require.NoError(suite.T(), err)
	return org.ParentOrgID
}

func (suite *OrgRouterTestSuite) TestMovingAnOrgUnderItsChildIsABadRequest() {
	w := serve(suite.routerFor(promptAdmin.roles...), http.MethodPut, orgPath(suite.org.ID)+"/parent", orgDTO.UpdateOrgParent{ParentOrgID: &suite.childOrg.ID})

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *OrgRouterTestSuite) TestDeletingAnOrgWithChildOrgsIsAConflict() {
	w := serve(suite.routerFor(promptAdmin.roles...), http.MethodDelete, orgPath(suite.org.ID), nil)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)
}

func (suite *OrgRouterTestSuite) TestUpdateOrgReturnsTheReplacedOrg() {
	w := serve(suite.routerFor(orgAdmin.roles...), http.MethodPut, orgPath(suite.org.ID), map[string]any{
		"name":    "Applied Software Engineering",
		"website": "https://aet.cit.tum.de",
		"slug":    "renamed",
	})

	require.Equal(suite.T(), http.StatusOK, w.Code)
	var updated orgDTO.Org
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(suite.T(), "Applied Software Engineering", updated.Name)
	assert.Equal(suite.T(), "https://aet.cit.tum.de", updated.Website.String)
	assert.Equal(suite.T(), "ase", updated.Slug, "a slug in the body is ignored")
}

func TestOrgRouterTestSuite(t *testing.T) {
	suite.Run(t, new(OrgRouterTestSuite))
}
