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

	// we need the keycloak middleware here to ensure that the user has a valid token
	teaseRouter.GET("/course-phases", keycloakTokenVerifier.KeycloakMiddleware(), getAllCoursePhases)

	// course phase specific endpoints
	teaseCoursePhaseRouter := teaseRouter.Group("/course_phase/:coursePhaseID")
	teaseCoursePhaseRouter.GET("/students", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTeaseStudentsForCoursePhase)
	teaseCoursePhaseRouter.GET("/skills", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTeaseSkillsByCoursePhase)
	teaseCoursePhaseRouter.GET("/projects", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTeaseTeamsByCoursePhase)

	teaseCoursePhaseRouter.POST("/allocations", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), postAllocations)
	teaseCoursePhaseRouter.GET("/allocations", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getAllocations)

	// Phase 1 Tease ↔ PROMPT workspace integration.
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

// getTeaseWorkspace handles GET /tease/course_phase/{coursePhaseID}/workspace.
// Returns 200 with the persisted workspace row, or 200 with empty defaults
// if no row exists yet (per §4.2 of the Phase 1 plan).
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

// putTeaseWorkspace handles PUT /tease/course_phase/{coursePhaseID}/workspace.
// Idempotent upsert. Does not touch the allocations table nor
// last_exported_at.
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

// postTeaseSave handles POST /tease/course_phase/{coursePhaseID}/save.
// Atomic: upsert tease_workspace + upsert allocations + stamp
// last_exported_at in one transaction. Invalid allocation payloads
// return 400; persistence errors roll back and return 500.
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

// getAuthenticatedUserID extracts the acting user's UUID from the
// token already parsed by the auth middleware. We never trust a
// client-supplied UpdatedBy for audit fields; the server stamps it
// from the bearer token.
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
