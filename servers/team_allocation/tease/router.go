package tease

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/team_allocation/tease/teaseDTO"
	log "github.com/sirupsen/logrus"
)

const maxTeaseWorkspaceBodyBytes = 5 << 20

func setupTeaseRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	teaseRouter := routerGroup.Group("/tease")

	teaseRouter.GET("/course-phases", keycloakTokenVerifier.KeycloakMiddleware(), getAllCoursePhases)

	teaseCoursePhaseRouter := teaseRouter.Group("/course_phase/:coursePhaseID")
	teaseCoursePhaseRouter.GET("/students", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTeaseStudentsForCoursePhase)
	teaseCoursePhaseRouter.GET("/skills", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTeaseSkillsByCoursePhase)
	teaseCoursePhaseRouter.GET("/projects", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTeaseTeamsByCoursePhase)

	teaseCoursePhaseRouter.POST("/allocations", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), postAllocations)
	teaseCoursePhaseRouter.GET("/allocations", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getAllocations)

	teaseCoursePhaseRouter.GET("/workspace", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTeaseWorkspace)
	teaseCoursePhaseRouter.PUT("/workspace", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), putTeaseWorkspace)
	teaseCoursePhaseRouter.POST("/save", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), postTeaseSave)
}

func getAllCoursePhases(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	rolesVal, exists := c.Get("userRoles")
	if !exists {
		handleError(c, http.StatusForbidden, errors.New("missing user roles"))
		return
	}

	userRoles, ok := rolesVal.(map[string]bool)
	if !ok {
		handleError(c, http.StatusInternalServerError, errors.New("invalid user roles format"))
		return
	}

	teasePhases, err := GetTeamAllocationCoursePhases(
		c,
		authHeader,
		userRoles,
	)

	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, teasePhases)
}

func getTeaseStudentsForCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}
	authHeader := c.GetHeader("Authorization")

	students, err := GetTeaseStudentsForCoursePhase(c, authHeader, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, students)
}

func getTeaseSkillsByCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	skills, err := GetTeaseSkillsByCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, skills)
}

func getTeaseTeamsByCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	teams, err := GetTeaseTeamsByCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, teams)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}

func getAllocations(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))

	if err != nil {
		handleError(c, http.StatusBadRequest, errors.New("invalid course phase ID"))
		return
	}

	allocations, err := GetAllocationsByCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, allocations)
}

func postAllocations(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, errors.New("invalid course phase ID"))
		return
	}

	var req []teaseDTO.Allocation
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = PostAllocations(c, req, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Allocations created successfully"})
}

// getTeaseWorkspace godoc
// @Summary Get TEASE workspace
// @Description Get the persisted TEASE workspace for a course phase
// @Tags tease
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} teaseDTO.TeaseWorkspace
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /tease/course_phase/{coursePhaseID}/workspace [get]
func getTeaseWorkspace(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, errors.New("invalid course phase ID"))
		return
	}

	workspace, err := GetTeaseWorkspace(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, workspace)
}

// putTeaseWorkspace godoc
// @Summary Save TEASE workspace draft
// @Description Upsert the TEASE workspace snapshot without publishing allocations
// @Tags tease
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body teaseDTO.TeaseWorkspaceRequest true "TEASE workspace"
// @Success 200 {object} teaseDTO.TeaseWorkspace
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 413 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /tease/course_phase/{coursePhaseID}/workspace [put]
func putTeaseWorkspace(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, errors.New("invalid course phase ID"))
		return
	}

	var req teaseDTO.TeaseWorkspaceRequest
	if !bindTeaseJSON(c, &req) {
		return
	}

	updatedBy, err := getAuthenticatedUserID(c)
	if err != nil {
		handleError(c, http.StatusUnauthorized, err)
		return
	}

	workspace, err := UpsertTeaseWorkspace(c, coursePhaseID, req, updatedBy)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, workspace)
}

// postTeaseSave godoc
// @Summary Publish TEASE allocations
// @Description Save the TEASE workspace and replace the published allocation set
// @Tags tease
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body teaseDTO.TeaseSaveRequest true "TEASE workspace and allocations"
// @Success 200 {object} teaseDTO.TeaseWorkspace
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 413 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /tease/course_phase/{coursePhaseID}/save [post]
func postTeaseSave(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, errors.New("invalid course phase ID"))
		return
	}

	var req teaseDTO.TeaseSaveRequest
	if !bindTeaseJSON(c, &req) {
		return
	}

	updatedBy, err := getAuthenticatedUserID(c)
	if err != nil {
		handleError(c, http.StatusUnauthorized, err)
		return
	}

	workspace, err := SaveTeaseWorkspaceAndAllocations(c, coursePhaseID, req, updatedBy)
	if err != nil {
		if errors.Is(err, errInvalidAllocation) {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func bindTeaseJSON(c *gin.Context, dst any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTeaseWorkspaceBodyBytes)
	if err := c.ShouldBindJSON(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			handleError(c, http.StatusRequestEntityTooLarge, fmt.Errorf("request body exceeds %d bytes", maxTeaseWorkspaceBodyBytes))
			return false
		}
		if errors.Is(err, io.EOF) {
			handleError(c, http.StatusBadRequest, errors.New("request body is required"))
			return false
		}
		handleError(c, http.StatusBadRequest, err)
		return false
	}
	return true
}

func getAuthenticatedUserID(c *gin.Context) (uuid.UUID, error) {
	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if !ok {
		return uuid.Nil, errors.New("user not found in context")
	}
	id, err := uuid.Parse(tokenUser.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
	}
	return id, nil
}
