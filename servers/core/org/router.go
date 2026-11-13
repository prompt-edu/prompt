package org

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/audit"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// @Summary Org Endpoints
// @Description Endpoints for managing orgs, the organizational units that own courses
// @Tags orgs
// @Security BearerAuth
func RegisterRoutes(router *gin.RouterGroup, service *OrgService, authMiddleware func() gin.HandlerFunc, checkOrgPermission permissionValidation.PermissionCheck) {
	setupOrgRouter(router, service, authMiddleware, permissionValidation.CheckAccessControlByRole, checkAccessControlByIDWrapper(checkOrgPermission))
}

func checkAccessControlByIDWrapper(check permissionValidation.PermissionCheck) func(allowedRoles ...string) gin.HandlerFunc {
	return func(allowedRoles ...string) gin.HandlerFunc {
		return permissionValidation.CheckAccessControlByID(check, "orgID", allowedRoles...)
	}
}

func setupOrgRouter(router *gin.RouterGroup, s *OrgService, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	org := router.Group("/orgs", authMiddleware())

	org.GET("/", permissionRoleMiddleware(permissionValidation.PromptAdmin), s.getAllOrgs)
	org.POST("/", audit.Describe("Created an org"), permissionRoleMiddleware(permissionValidation.PromptAdmin), s.createOrg)
	// Every caller may list their own orgs; the token's org roles decide which.
	org.GET("/self", s.getOwnOrgs)
	org.GET("/:orgID", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.OrgAdmin, permissionValidation.OrgMember), s.getOrg)
	org.PUT("/:orgID", audit.Describe("Updated an org"), permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.OrgAdmin), s.updateOrg)
	org.PUT("/:orgID/parent", audit.Describe("Moved an org"), permissionRoleMiddleware(permissionValidation.PromptAdmin), s.updateOrgParent)
	org.DELETE("/:orgID", audit.Describe("Deleted an org"), permissionRoleMiddleware(permissionValidation.PromptAdmin), s.deleteOrg)

	org.GET("/:orgID/members", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.OrgAdmin), s.getOrgMembers)
	org.PUT("/:orgID/members", audit.Describe("Set the role of an org member"), permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.OrgAdmin), s.setOrgMember)
	org.PUT("/:orgID/members/:userID", audit.Describe("Changed the role of an org member"), permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.OrgAdmin), s.updateOrgMemberRole)
	org.DELETE("/:orgID/members/:userID", audit.Describe("Removed an org member"), permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.OrgAdmin), s.removeOrgMember)
}

// getOwnOrgs godoc
// @Summary Get the caller's orgs
// @Description Get the orgs whose Admin or Member role the caller's token holds, each with the caller's role in it. A new role shows only after a token refresh.
// @Tags orgs
// @Produce json
// @Success 200 {array} orgDTO.OwnOrg
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/self [get]
func (s *OrgService) getOwnOrgs(c *gin.Context) {
	userRoles, ok := c.Value(keycloakTokenVerifier.CtxUserRoles).(map[string]bool)
	if !ok {
		log.Error("user roles not found in context")
		sdkUtils.HandleError(c, http.StatusInternalServerError, errors.New("failed to get own orgs"))
		return
	}
	issuedAt, err := permissionValidation.TokenIssuedAt(c)
	if err != nil {
		log.Error(err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, errors.New("failed to get own orgs"))
		return
	}

	orgs, err := s.GetOwnOrgs(c, userRoles, issuedAt)
	if err != nil {
		log.Error(err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, errors.New("failed to get own orgs"))
		return
	}
	c.IndentedJSON(http.StatusOK, orgs)
}

// getOrgMembers godoc
// @Summary Get the members of an org
// @Description Get the users in the org's Admin and Member groups, Admins first. A user in both groups is listed once, as an Admin.
// @Tags orgs
// @Produce json
// @Param orgID path string true "Org UUID"
// @Success 200 {array} keycloakRealmDTO.OrgMember
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgID}/members [get]
func (s *OrgService) getOrgMembers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	members, err := s.GetOrgMembers(c, id)
	if err != nil {
		handleServiceError(c, err, "failed to get org members")
		return
	}
	c.IndentedJSON(http.StatusOK, members)
}

// setOrgMember godoc
// @Summary Set the role of an org member
// @Description Give the user with the university login the Admin or Member role of the org, adding them to the org if they are not a member yet. A user holds one org role at a time. Callers cannot change their own membership.
// @Tags orgs
// @Accept json
// @Produce json
// @Param orgID path string true "Org UUID"
// @Param member body orgDTO.SetOrgMember true "University login and org role"
// @Success 200 {object} keycloakRealmDTO.OrgMember
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgID}/members [put]
func (s *OrgService) setOrgMember(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var input orgDTO.SetOrgMember
	if err := c.ShouldBindJSON(&input); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}
	if err := validateSetOrgMember(input); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}
	callerUserID, ok := callerID(c)
	if !ok {
		return
	}

	org, member, err := s.SetOrgMember(c, id, input, callerUserID)
	if err != nil {
		handleServiceError(c, err, "failed to set the org member's role")
		return
	}
	recordOrgMemberAudit(c, org, member.KeycloakUserID, member.Role)
	c.IndentedJSON(http.StatusOK, member)
}

// updateOrgMemberRole godoc
// @Summary Change the role of an org member
// @Description Give the user with the Keycloak ID the Admin or Member role of the org, adding them to the org if they are not a member yet. A user holds one org role at a time. Callers cannot change their own membership.
// @Tags orgs
// @Accept json
// @Param orgID path string true "Org UUID"
// @Param userID path string true "Keycloak user ID"
// @Param role body orgDTO.UpdateOrgMemberRole true "Org role"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgID}/members/{userID} [put]
func (s *OrgService) updateOrgMemberRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}
	userID, ok := memberID(c)
	if !ok {
		return
	}

	var input orgDTO.UpdateOrgMemberRole
	if err := c.ShouldBindJSON(&input); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}
	if err := validateOrgRole(input.Role); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}
	callerUserID, ok := callerID(c)
	if !ok {
		return
	}

	org, err := s.UpdateOrgMemberRole(c, id, userID, input.Role, callerUserID)
	if err != nil {
		handleServiceError(c, err, "failed to change the org member's role")
		return
	}
	recordOrgMemberAudit(c, org, userID, input.Role)
	c.Status(http.StatusNoContent)
}

// removeOrgMember godoc
// @Summary Remove a member from an org
// @Description Take the user out of the org's Admin and Member groups. Removing a user who is not a member changes nothing. Callers cannot remove themselves.
// @Tags orgs
// @Param orgID path string true "Org UUID"
// @Param userID path string true "Keycloak user ID"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgID}/members/{userID} [delete]
func (s *OrgService) removeOrgMember(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}
	userID, ok := memberID(c)
	if !ok {
		return
	}
	callerUserID, ok := callerID(c)
	if !ok {
		return
	}

	org, err := s.RemoveOrgMember(c, id, userID, callerUserID)
	if err != nil {
		handleServiceError(c, err, "failed to remove the org member")
		return
	}
	recordOrgMemberAudit(c, org, userID, "")
	c.Status(http.StatusNoContent)
}

// memberID returns the Keycloak user ID in the URL in its canonical form, or refuses the request.
// Keycloak user IDs are UUIDs, so only a well-formed one is passed on to Keycloak. The canonical
// form also keeps the own-membership guard from missing an uppercase spelling of the caller's ID.
func memberID(c *gin.Context) (string, bool) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return "", false
	}
	return userID.String(), true
}

// callerID returns the caller's Keycloak user ID. Without it, the own-membership guard would not
// hold, so the request is refused.
func callerID(c *gin.Context) (string, bool) {
	callerUserID := c.GetString(keycloakTokenVerifier.CtxUserID)
	if callerUserID == "" {
		sdkUtils.HandleError(c, http.StatusUnauthorized, errors.New("missing caller identity"))
		return "", false
	}
	return callerUserID, true
}

// getAllOrgs godoc
// @Summary Get all orgs
// @Description Get every org as a flat list; the hierarchy is given by parentOrgID
// @Tags orgs
// @Produce json
// @Success 200 {array} orgDTO.Org
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/ [get]
func (s *OrgService) getAllOrgs(c *gin.Context) {
	orgs, err := s.GetAllOrgs(c)
	if err != nil {
		log.Error(err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, errors.New("failed to get orgs"))
		return
	}
	c.IndentedJSON(http.StatusOK, orgs)
}

// getOrg godoc
// @Summary Get an org
// @Description Get an org by ID
// @Tags orgs
// @Produce json
// @Param orgID path string true "Org UUID"
// @Success 200 {object} orgDTO.Org
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgID} [get]
func (s *OrgService) getOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	org, err := s.GetOrg(c, id)
	if err != nil {
		handleServiceError(c, err, "failed to get org")
		return
	}
	c.IndentedJSON(http.StatusOK, org)
}

// createOrg godoc
// @Summary Create an org
// @Description Create an org, optionally under a parent org, and provision its Keycloak groups and roles
// @Tags orgs
// @Accept json
// @Produce json
// @Param org body orgDTO.CreateOrg true "Org to create"
// @Success 201 {object} orgDTO.Org
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/ [post]
func (s *OrgService) createOrg(c *gin.Context) {
	var input orgDTO.CreateOrg
	if err := c.ShouldBindJSON(&input); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}
	if err := validateCreateOrg(input); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	org, err := s.CreateOrg(c, input)
	if err != nil {
		handleServiceError(c, err, "failed to create org")
		return
	}
	recordOrgAudit(c, org.ID, org.Slug)
	c.IndentedJSON(http.StatusCreated, org)
}

// updateOrg godoc
// @Summary Update an org
// @Description Replace the name and the optional fields of an org; an empty optional field clears it. The slug and the parent org cannot be changed here.
// @Tags orgs
// @Accept json
// @Produce json
// @Param orgID path string true "Org UUID"
// @Param org body orgDTO.UpdateOrg true "Updated org"
// @Success 200 {object} orgDTO.Org
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgID} [put]
func (s *OrgService) updateOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var input orgDTO.UpdateOrg
	if err := c.ShouldBindJSON(&input); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}
	if err := validateUpdateOrg(input); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	org, err := s.UpdateOrg(c, id, input)
	if err != nil {
		handleServiceError(c, err, "failed to update org")
		return
	}
	recordOrgAudit(c, org.ID, org.Slug)
	c.IndentedJSON(http.StatusOK, org)
}

// updateOrgParent godoc
// @Summary Move an org
// @Description Place an org under another org, or at the top level when parentOrgID is null. An org cannot be placed under itself or one of its descendants.
// @Tags orgs
// @Accept json
// @Produce json
// @Param orgID path string true "Org UUID"
// @Param parent body orgDTO.UpdateOrgParent true "New parent org"
// @Success 200 {object} orgDTO.Org
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgID}/parent [put]
func (s *OrgService) updateOrgParent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var input orgDTO.UpdateOrgParent
	if err := c.ShouldBindJSON(&input); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	org, err := s.UpdateOrgParent(c, id, input)
	if err != nil {
		handleServiceError(c, err, "failed to move org")
		return
	}
	recordOrgAudit(c, org.ID, org.Slug)
	c.IndentedJSON(http.StatusOK, org)
}

// deleteOrg godoc
// @Summary Delete an org
// @Description Delete an org without child orgs or courses, together with its Keycloak groups and roles
// @Tags orgs
// @Param orgID path string true "Org UUID"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgID} [delete]
func (s *OrgService) deleteOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	slug, err := s.DeleteOrg(c, id)
	if err != nil {
		handleServiceError(c, err, "failed to delete org")
		return
	}
	recordOrgAudit(c, id, slug)
	c.Status(http.StatusNoContent)
}

func recordOrgAudit(c *gin.Context, orgID uuid.UUID, slug string) {
	audit.Record(c, audit.Event{EntityType: "org", EntityID: orgID.String(), EntityName: slug})
}

// recordOrgMemberAudit names the org and the member; role is empty for a removal.
func recordOrgMemberAudit(c *gin.Context, org orgDTO.Org, userID, role string) {
	metadata := map[string]any{"keycloakUserID": userID}
	if role != "" {
		metadata["role"] = role
	}
	audit.Record(c, audit.Event{EntityType: "org", EntityID: org.ID.String(), EntityName: org.Slug, Metadata: metadata})
}

func handleServiceError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, ErrOrgNotFound):
		sdkUtils.HandleError(c, http.StatusNotFound, err)
	case errors.Is(err, keycloakRealmManager.ErrUniversityLoginNotFound), errors.Is(err, keycloakRealmManager.ErrUserNotFound):
		sdkUtils.HandleError(c, http.StatusNotFound, err)
	case errors.Is(err, ErrParentOrgNotFound), errors.Is(err, ErrOrgCycle), errors.Is(err, ErrOwnMembership), errors.Is(err, keycloakRealmManager.ErrInvalidOrgRole):
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
	case errors.Is(err, ErrDuplicateOrgSlug), errors.Is(err, ErrOrgHasChildOrgs), errors.Is(err, ErrOrgHasCourses), errors.Is(err, keycloakRealmManager.ErrAmbiguousUniversityLogin):
		sdkUtils.HandleError(c, http.StatusConflict, err)
	case errors.Is(err, ErrOrgMembersRemoved):
		log.Error(err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, ErrOrgMembersRemoved)
	default:
		log.Error(err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, errors.New(fallbackMessage))
	}
}
