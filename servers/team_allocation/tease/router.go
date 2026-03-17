package tease

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/team_allocation/tease/teaseDTO"
	log "github.com/sirupsen/logrus"
)

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
