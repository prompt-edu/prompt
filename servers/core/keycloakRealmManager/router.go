package keycloakRealmManager

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

// setupKeycloakRouter sets up the keycloak endpoints
// @Summary Keycloak Endpoints
// @Description Endpoints for managing Keycloak groups and students
// @Tags keycloak
// @Security BearerAuth
func setupKeycloakRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	keycloak := router.Group("/keycloak/:courseID", authMiddleware())
	keycloak.PUT("/group", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), createCustomGroup)
	keycloak.GET("/group/:groupName/students", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), getStudentsInGroup)
	// Adding Students to the editor role of a course
	keycloak.PUT("/group/editor/students", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), addStudentsToEditorGroup)
	// Adding Students to a custom group
	keycloak.PUT("/group/:groupName/students", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), addStudentsToGroup)
}

// createCustomGroup godoc
// @Summary Create a custom group
// @Description Create a new custom group for a course
// @Tags keycloak
// @Accept json
// @Produce json
// @Param courseID path string true "Course UUID"
// @Param newGroupName body keycloakRealmDTO.CreateGroup true "Group name to create"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /keycloak/{courseID}/group [put]
func createCustomGroup(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("courseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var newGroupName keycloakRealmDTO.CreateGroup
	if err := c.BindJSON(&newGroupName); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	id, err := AddCustomGroup(c, courseID, newGroupName.GroupName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"id": id})
}

// addStudentsToGroup godoc
// @Summary Add students to a custom group
// @Description Add students to a custom group for a course
// @Tags keycloak
// @Accept json
// @Produce json
// @Param courseID path string true "Course UUID"
// @Param groupName path string true "Group name"
// @Param request body keycloakRealmDTO.AddStudentsToGroup true "Students to add"
// @Success 200 {object} keycloakRealmDTO.AddStudentsToGroupResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /keycloak/{courseID}/group/{groupName}/students [put]
func addStudentsToGroup(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("courseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	groupName := c.Param("groupName")
	if groupName == "" {
		handleError(c, http.StatusBadRequest, errors.New("group name is required"))
		return
	}

	var request keycloakRealmDTO.AddStudentsToGroup
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	addingReport, err := AddStudentsToGroup(c, courseID, request.StudentsToAdd, groupName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, addingReport)
}

// addStudentsToEditorGroup godoc
// @Summary Add students to the editor group
// @Description Add students to the editor group for a course
// @Tags keycloak
// @Accept json
// @Produce json
// @Param courseID path string true "Course UUID"
// @Param request body keycloakRealmDTO.AddStudentsToGroup true "Students to add"
// @Success 200 {object} keycloakRealmDTO.AddStudentsToGroupResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /keycloak/{courseID}/group/editor/students [put]
func addStudentsToEditorGroup(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("courseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request keycloakRealmDTO.AddStudentsToGroup
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	addingReport, err := AddStudentsToEditorGroup(c, courseID, request.StudentsToAdd)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, addingReport)
}

// getStudentsInGroup godoc
// @Summary Get students in a group
// @Description Get all students in a specific group for a course
// @Tags keycloak
// @Produce json
// @Param courseID path string true "Course UUID"
// @Param groupName path string true "Group name"
// @Success 200 {array} keycloakRealmDTO.GroupMembers
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /keycloak/{courseID}/group/{groupName}/students [get]
func getStudentsInGroup(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("courseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	groupName := c.Param("groupName")
	if groupName == "" {
		handleError(c, http.StatusBadRequest, errors.New("group name is required"))
		return
	}

	students, err := GetStudentsInGroup(c, courseID, groupName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, students)

}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
