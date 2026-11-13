package org

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/audit"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OrgRouterTestSuite struct {
	suite.Suite
	ctx               context.Context
	cleanup           func()
	service           *OrgService
	validationService *permissionValidation.ValidationService
	members           *fakeMemberDirectory
	org               orgDTO.Org
	childOrg          orgDTO.Org
	otherOrg          orgDTO.Org
}

// callerUserID is the Keycloak user ID every test request is made as.
const callerUserID = "caller"

func (suite *OrgRouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	noKeycloak := func(context.Context, string) error { return nil }
	noKeycloakDelete := func(context.Context, string) (orgMembers, error) { return nil, nil }
	noKeycloakArtifacts := func(context.Context, string) (bool, error) { return false, nil }
	noKeycloakMembers := func(context.Context, string, orgMembers) error { return nil }
	suite.members = newFakeMemberDirectory(staffMember("ada", "ad12min"), staffMember(callerUserID, "ca11ler"))
	suite.service = NewOrgService(*testDB.Queries, testDB.Conn, noKeycloak, noKeycloakDelete, noKeycloakArtifacts, noKeycloakMembers, suite.members)
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
	return suite.routerWith(roles)
}

func (suite *OrgRouterTestSuite) routerWith(roles []string, middleware ...gin.HandlerFunc) *gin.Engine {
	// Issued after every org the tests create, even if the database clock runs slightly ahead.
	return suite.routerForTokenIssuedAt(time.Now().Add(time.Minute), roles, middleware...)
}

func (suite *OrgRouterTestSuite) routerForTokenIssuedAt(issuedAt time.Time, roles []string, middleware ...gin.HandlerFunc) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api", middleware...)
	RegisterRoutes(api, suite.service, func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(keycloakTokenVerifier.CtxTokenIssuedAt, issuedAt)
			c.Set(keycloakTokenVerifier.CtxUserID, callerUserID)
			sdkTestUtils.MockAuthMiddleware(roles)(c)
		}
	}, suite.validationService.CheckOrgPermission)
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

func (suite *OrgRouterTestSuite) TestRolesOfADeletedOrgGrantNothingOnAnOrgReusingItsSlug() {
	deleted := suite.createOrg("reused", nil)
	_, err := suite.service.DeleteOrg(suite.ctx, deleted.ID)
	require.NoError(suite.T(), err)
	reused := suite.createOrg("reused", nil)
	roles := []string{"org-reused-Admin"}

	staleToken := suite.routerForTokenIssuedAt(reused.CreatedAt.Add(-time.Minute), roles)
	sameSecondToken := suite.routerForTokenIssuedAt(reused.CreatedAt.Truncate(time.Second), roles)
	freshToken := suite.routerForTokenIssuedAt(reused.CreatedAt.Add(time.Second), roles)

	assert.Equal(suite.T(), http.StatusForbidden, serve(staleToken, http.MethodGet, orgPath(reused.ID), nil).Code)
	assert.Equal(suite.T(), http.StatusForbidden, serve(sameSecondToken, http.MethodGet, orgPath(reused.ID), nil).Code)
	assert.Equal(suite.T(), http.StatusOK, serve(freshToken, http.MethodGet, orgPath(reused.ID), nil).Code)
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

type auditCapture chan audit.Event

func (a auditCapture) Record(_ context.Context, e audit.Event) error {
	a <- e
	return nil
}

func (suite *OrgRouterTestSuite) TestAuditEntriesNameTheOrg() {
	events := make(auditCapture, 1)
	router := suite.routerWith(promptAdmin.roles, audit.Middleware(events, audit.WithActorExtractor(func(*gin.Context) (audit.Actor, bool) {
		return audit.Actor{ID: "admin"}, true
	})))
	slug := "audited-" + uuid.NewString()[:8]

	w := serve(router, http.MethodPost, "/api/orgs/", orgDTO.CreateOrg{Name: "Audited Org", Slug: slug})
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	var created orgDTO.Org
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &created))
	suite.expectAuditEvent(events, "Created an org", created.ID, slug)

	for _, request := range []struct {
		method string
		path   string
		body   any
		status int
		action string
		name   string
	}{
		{http.MethodPut, orgPath(created.ID), orgDTO.UpdateOrg{Name: "Audited Org Renamed"}, http.StatusOK, "Updated an org", slug},
		{http.MethodPut, orgPath(created.ID) + "/parent", orgDTO.UpdateOrgParent{ParentOrgID: &suite.otherOrg.ID}, http.StatusOK, "Moved an org", slug},
		{http.MethodDelete, orgPath(created.ID), nil, http.StatusNoContent, "Deleted an org", slug},
	} {
		w := serve(router, request.method, request.path, request.body)
		require.Equal(suite.T(), request.status, w.Code, request.action)
		suite.expectAuditEvent(events, request.action, created.ID, request.name)
	}
}

func (suite *OrgRouterTestSuite) expectAuditEvent(events auditCapture, action string, orgID uuid.UUID, name string) audit.Event {
	select {
	case event := <-events:
		assert.Equal(suite.T(), action, event.Action)
		assert.Equal(suite.T(), "org", event.EntityType, action)
		assert.Equal(suite.T(), orgID.String(), event.EntityID, action)
		assert.Equal(suite.T(), name, event.EntityName, action)
		return event
	case <-time.After(5 * time.Second):
		suite.T().Fatalf("no audit event for %q", action)
		return audit.Event{}
	}
}

func (suite *OrgRouterTestSuite) TestMalformedBodiesGetAJSONError() {
	for _, request := range []struct{ method, path string }{
		{http.MethodPost, "/api/orgs/"},
		{http.MethodPut, orgPath(suite.org.ID)},
		{http.MethodPut, orgPath(suite.org.ID) + "/parent"},
		{http.MethodPut, membersPath(suite.org.ID)},
	} {
		req, _ := http.NewRequest(request.method, request.path, strings.NewReader(`{"name":`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.routerFor(promptAdmin.roles...).ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code, request.path)
		assert.Contains(suite.T(), w.Result().Header.Get("Content-Type"), "application/json", request.path)
	}
}

func membersPath(orgID uuid.UUID) string {
	return orgPath(orgID) + "/members"
}

func (suite *OrgRouterTestSuite) TestListOrgMembersPermissions() {
	suite.expectStatus(http.MethodGet, func() string { return membersPath(suite.org.ID) }, noBody, http.StatusOK, promptAdmin, orgAdmin)
}

func (suite *OrgRouterTestSuite) TestSetOrgMemberPermissions() {
	suite.expectStatus(http.MethodPut, func() string { return membersPath(suite.org.ID) }, func() any {
		return orgDTO.SetOrgMember{UniversityLogin: "ad12min", Role: permissionValidation.OrgMember}
	}, http.StatusOK, promptAdmin, orgAdmin)
}

func (suite *OrgRouterTestSuite) TestRemoveOrgMemberPermissions() {
	suite.expectStatus(http.MethodDelete, func() string { return membersPath(suite.org.ID) + "/ada" }, noBody, http.StatusNoContent, promptAdmin, orgAdmin)
}

func (suite *OrgRouterTestSuite) TestParentOrgAdminsCannotManageTheMembersOfChildOrgs() {
	suite.expectStatus(http.MethodGet, func() string { return membersPath(suite.childOrg.ID) }, noBody, http.StatusOK, promptAdmin)
	suite.expectStatus(http.MethodPut, func() string { return membersPath(suite.childOrg.ID) }, func() any {
		return orgDTO.SetOrgMember{UniversityLogin: "ad12min", Role: permissionValidation.OrgAdmin}
	}, http.StatusOK, promptAdmin)
	suite.expectStatus(http.MethodDelete, func() string { return membersPath(suite.childOrg.ID) + "/ada" }, noBody, http.StatusNoContent, promptAdmin)
}

func (suite *OrgRouterTestSuite) TestSetOrgMemberGivesTheUserTheRole() {
	router := suite.routerFor(orgAdmin.roles...)

	w := serve(router, http.MethodPut, membersPath(suite.org.ID), orgDTO.SetOrgMember{UniversityLogin: "ad12min", Role: permissionValidation.OrgAdmin})
	require.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), permissionValidation.OrgAdmin, suite.members.roleOf("ase", "ada"))

	w = serve(router, http.MethodPut, membersPath(suite.org.ID), orgDTO.SetOrgMember{UniversityLogin: "ad12min", Role: permissionValidation.OrgMember})
	require.Equal(suite.T(), http.StatusOK, w.Code)
	assert.JSONEq(suite.T(), `{
		"keycloakUserID": "ada",
		"username": "ada",
		"email": "",
		"firstName": "",
		"lastName": "",
		"universityLogin": "ad12min",
		"role": "Member"
	}`, w.Body.String())
	assert.Equal(suite.T(), permissionValidation.OrgMember, suite.members.roleOf("ase", "ada"))
}

func (suite *OrgRouterTestSuite) TestSetOrgMemberValidatesInput() {
	for _, tc := range []struct {
		name   string
		body   orgDTO.SetOrgMember
		status int
	}{
		{"blank university login", orgDTO.SetOrgMember{UniversityLogin: "  ", Role: permissionValidation.OrgMember}, http.StatusBadRequest},
		{"university login with a space", orgDTO.SetOrgMember{UniversityLogin: "ad 12min", Role: permissionValidation.OrgMember}, http.StatusBadRequest},
		{"university login with a colon", orgDTO.SetOrgMember{UniversityLogin: "ad:12min", Role: permissionValidation.OrgMember}, http.StatusBadRequest},
		{"university login too long", orgDTO.SetOrgMember{UniversityLogin: strings.Repeat("a", 256), Role: permissionValidation.OrgMember}, http.StatusBadRequest},
		{"no role", orgDTO.SetOrgMember{UniversityLogin: "ad12min"}, http.StatusBadRequest},
		{"lowercase role", orgDTO.SetOrgMember{UniversityLogin: "ad12min", Role: "admin"}, http.StatusBadRequest},
		{"course role", orgDTO.SetOrgMember{UniversityLogin: "ad12min", Role: permissionValidation.CourseLecturer}, http.StatusBadRequest},
		{"unknown university login", orgDTO.SetOrgMember{UniversityLogin: "un12own", Role: permissionValidation.OrgMember}, http.StatusNotFound},
	} {
		w := serve(suite.routerFor(orgAdmin.roles...), http.MethodPut, membersPath(suite.org.ID), tc.body)

		assert.Equal(suite.T(), tc.status, w.Code, tc.name)
		assert.True(suite.T(), json.Valid(w.Body.Bytes()), tc.name)
	}
}

func (suite *OrgRouterTestSuite) TestNobodyChangesTheirOwnMembership() {
	require.NoError(suite.T(), suite.members.SetOrgMemberRole(suite.ctx, "ase", callerUserID, permissionValidation.OrgAdmin))
	router := suite.routerFor(orgAdmin.roles...)

	demote := serve(router, http.MethodPut, membersPath(suite.org.ID), orgDTO.SetOrgMember{UniversityLogin: "ca11ler", Role: permissionValidation.OrgMember})
	leave := serve(router, http.MethodDelete, membersPath(suite.org.ID)+"/"+callerUserID, nil)

	assert.Equal(suite.T(), http.StatusBadRequest, demote.Code)
	assert.Equal(suite.T(), http.StatusBadRequest, leave.Code)
	assert.Equal(suite.T(), permissionValidation.OrgAdmin, suite.members.roleOf("ase", callerUserID))
}

func (suite *OrgRouterTestSuite) TestRemovingAnUnknownUserIsNotFound() {
	w := serve(suite.routerFor(orgAdmin.roles...), http.MethodDelete, membersPath(suite.org.ID)+"/ghost", nil)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *OrgRouterTestSuite) TestMembersOfAnUnknownOrgAreNotFoundForAdminsOnly() {
	path := membersPath(uuid.New())

	assert.Equal(suite.T(), http.StatusNotFound, serve(suite.routerFor(promptAdmin.roles...), http.MethodGet, path, nil).Code)
	assert.Equal(suite.T(), http.StatusForbidden, serve(suite.routerFor(orgAdmin.roles...), http.MethodGet, path, nil).Code)
}

func (suite *OrgRouterTestSuite) TestMembershipErrorsFromKeycloakAreServerErrors() {
	suite.members.err = errors.New("keycloak is down")
	defer func() { suite.members.err = nil }()

	w := serve(suite.routerFor(orgAdmin.roles...), http.MethodGet, membersPath(suite.org.ID), nil)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.NotContains(suite.T(), w.Body.String(), "keycloak is down")
}

func (suite *OrgRouterTestSuite) TestOwnOrgsFollowTheTokensOrgRoles() {
	for _, tc := range []struct {
		actor actor
		want  map[uuid.UUID]string
	}{
		{promptAdmin, map[uuid.UUID]string{}},
		{orgAdmin, map[uuid.UUID]string{suite.org.ID: permissionValidation.OrgAdmin}},
		{orgMember, map[uuid.UUID]string{suite.org.ID: permissionValidation.OrgMember}},
		{otherAdmin, map[uuid.UUID]string{suite.otherOrg.ID: permissionValidation.OrgAdmin}},
		{customGroup, map[uuid.UUID]string{}},
		{lecturer, map[uuid.UUID]string{}},
		{noOrgRoles, map[uuid.UUID]string{}},
	} {
		w := serve(suite.routerFor(tc.actor.roles...), http.MethodGet, "/api/orgs/self", nil)

		require.Equal(suite.T(), http.StatusOK, w.Code, tc.actor.name)
		var ownOrgs []orgDTO.OwnOrg
		require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &ownOrgs), tc.actor.name)
		got := map[uuid.UUID]string{}
		for _, ownOrg := range ownOrgs {
			got[ownOrg.ID] = ownOrg.Role
		}
		assert.Equal(suite.T(), tc.want, got, tc.actor.name)
	}
}

func (suite *OrgRouterTestSuite) TestMembershipAuditEntriesNameTheOrgAndTheMember() {
	events := make(auditCapture, 1)
	router := suite.routerWith(orgAdmin.roles, audit.Middleware(events, audit.WithActorExtractor(func(*gin.Context) (audit.Actor, bool) {
		return audit.Actor{ID: callerUserID}, true
	})))

	w := serve(router, http.MethodPut, membersPath(suite.org.ID), orgDTO.SetOrgMember{UniversityLogin: "ad12min", Role: permissionValidation.OrgAdmin})
	require.Equal(suite.T(), http.StatusOK, w.Code)
	set := suite.expectAuditEvent(events, "Set the role of an org member", suite.org.ID, "ase")
	assert.Equal(suite.T(), map[string]any{"keycloakUserID": "ada", "role": permissionValidation.OrgAdmin}, set.Metadata)

	w = serve(router, http.MethodDelete, membersPath(suite.org.ID)+"/ada", nil)
	require.Equal(suite.T(), http.StatusNoContent, w.Code)
	removed := suite.expectAuditEvent(events, "Removed an org member", suite.org.ID, "ase")
	assert.Equal(suite.T(), map[string]any{"keycloakUserID": "ada"}, removed.Metadata)
}

func TestOrgRouterTestSuite(t *testing.T) {
	suite.Run(t, new(OrgRouterTestSuite))
}
