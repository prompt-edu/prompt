package org

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/audit"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes mounts the org endpoints on the given router group.
// @Summary Org Endpoints
// @Description Endpoints for managing orgs, the organizational units that own courses
// @Tags orgs
// @Security BearerAuth
func RegisterRoutes(router *gin.RouterGroup, service *OrgService, authMiddleware func() gin.HandlerFunc, checkOrgPermission permissionValidation.PermissionCheck) {
	setupOrgRouter(router, service, authMiddleware, permissionValidation.CheckAccessControlByRole, checkAccessControlByIDWrapper(checkOrgPermission))
}

// initializes the handler func with CheckOrgPermission
func checkAccessControlByIDWrapper(check permissionValidation.PermissionCheck) func(allowedRoles ...string) gin.HandlerFunc {
	return func(allowedRoles ...string) gin.HandlerFunc {
		return permissionValidation.CheckAccessControlByID(check, "orgID", allowedRoles...)
	}
}

func setupOrgRouter(router *gin.RouterGroup, s *OrgService, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	org := router.Group("/orgs", authMiddleware())

	org.GET("/", permissionRoleMiddleware(permissionValidation.PromptAdmin), s.getAllOrgs)
	org.POST("/", audit.Describe("Created an org"), permissionRoleMiddleware(permissionValidation.PromptAdmin), s.createOrg)
	org.GET("/:orgID", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.OrgAdmin, permissionValidation.OrgMember), s.getOrg)
	org.PUT("/:orgID", audit.Describe("Updated an org"), permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.OrgAdmin), s.updateOrg)
	org.PUT("/:orgID/parent", audit.Describe("Moved an org"), permissionRoleMiddleware(permissionValidation.PromptAdmin), s.updateOrgParent)
	org.DELETE("/:orgID", audit.Describe("Deleted an org"), permissionRoleMiddleware(permissionValidation.PromptAdmin), s.deleteOrg)
}

// getAllOrgs godoc
// @Summary Get all orgs
// @Description Get every org as a flat list; the hierarchy is given by parentOrgID
// @Tags orgs
// @Produce json
// @Success 200 {array} orgDTO.Org
// @Failure 500 {object} utils.ErrorResponse
// @Router /orgs/ [get]
func (s *OrgService) getAllOrgs(c *gin.Context) {
	orgs, err := s.GetAllOrgs(c)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to get orgs"))
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
// @Router /orgs/{orgID} [get]
func (s *OrgService) getOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
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
// @Router /orgs/ [post]
func (s *OrgService) createOrg(c *gin.Context) {
	var input orgDTO.CreateOrg
	if err := c.BindJSON(&input); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := validateCreateOrg(input); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	org, err := s.CreateOrg(c, input)
	if err != nil {
		handleServiceError(c, err, "failed to create org")
		return
	}
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
// @Router /orgs/{orgID} [put]
func (s *OrgService) updateOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var input orgDTO.UpdateOrg
	if err := c.BindJSON(&input); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := validateUpdateOrg(input); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	org, err := s.UpdateOrg(c, id, input)
	if err != nil {
		handleServiceError(c, err, "failed to update org")
		return
	}
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
// @Router /orgs/{orgID}/parent [put]
func (s *OrgService) updateOrgParent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var input orgDTO.UpdateOrgParent
	if err := c.BindJSON(&input); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	org, err := s.UpdateOrgParent(c, id, input)
	if err != nil {
		handleServiceError(c, err, "failed to move org")
		return
	}
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
// @Router /orgs/{orgID} [delete]
func (s *OrgService) deleteOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := s.DeleteOrg(c, id); err != nil {
		handleServiceError(c, err, "failed to delete org")
		return
	}
	c.Status(http.StatusNoContent)
}

func handleServiceError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, ErrOrgNotFound):
		handleError(c, http.StatusNotFound, err)
	case errors.Is(err, ErrParentOrgNotFound), errors.Is(err, ErrOrgCycle):
		handleError(c, http.StatusBadRequest, err)
	case errors.Is(err, ErrDuplicateOrgSlug), errors.Is(err, ErrOrgHasChildOrgs), errors.Is(err, ErrOrgHasCourses):
		handleError(c, http.StatusConflict, err)
	default:
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New(fallbackMessage))
	}
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, sdkUtils.ErrorResponse{
		Error: err.Error(),
	})
}
