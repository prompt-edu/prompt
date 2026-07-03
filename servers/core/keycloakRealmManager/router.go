package keycloakRealmManager

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
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

	// Course staff management (Lecturer & Editor membership)
	keycloak.GET("/group/staff", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), getCourseStaff)
	keycloak.PUT("/group/:groupName/members/:userID", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), addCourseStaffMember)
	keycloak.DELETE("/group/:groupName/members/:userID", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), removeCourseStaffMember)

	// Realm-wide Keycloak user search. Not nested under :courseID because the
	// query has no course context; permissionValidation.CheckAccessControlByRole
	// matches CourseLecturer against any role suffix, so "is a lecturer of some
	// course" passes without needing a courseID. Don't refactor this back under
	// :courseID without re-checking that comment.
	realmWide := router.Group("/keycloak", authMiddleware())
	realmWide.GET("/users/search",
		permissionValidation.CheckAccessControlByRole(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer),
		searchKeycloakUsers,
	)
	// Service-account health probe for the admin system-status page.
	realmWide.GET("/status",
		permissionValidation.CheckAccessControlByRole(permissionValidation.PromptAdmin),
		getKeycloakStatus,
	)
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

// getCourseStaff godoc
// @Summary Get the staff (Lecturers and Editors) of a course
// @Tags keycloak
// @Produce json
// @Param courseID path string true "Course UUID"
// @Success 200 {object} keycloakRealmDTO.CourseStaff
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /keycloak/{courseID}/group/staff [get]
func getCourseStaff(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("courseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	staff, err := GetCourseStaff(c, courseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, staff)
}

// addCourseStaffMember godoc
// @Summary Add a Keycloak user to the Lecturer or Editor group of a course
// @Tags keycloak
// @Produce json
// @Param courseID path string true "Course UUID"
// @Param groupName path string true "Group name (Lecturer or Editor)"
// @Param userID path string true "Keycloak user ID"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /keycloak/{courseID}/group/{groupName}/members/{userID} [put]
func addCourseStaffMember(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("courseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	groupName := c.Param("groupName")
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		handleError(c, http.StatusBadRequest, errors.New("userID is required"))
		return
	}

	callerUserID := c.GetString(keycloakTokenVerifier.CtxUserID)
	if callerUserID == "" {
		// Defence in depth: the auth middleware should always set the caller's
		// Keycloak `sub` here. If it didn't, refuse the mutation rather than
		// silently writing audit lines with an empty caller.
		handleError(c, http.StatusUnauthorized, errors.New("missing caller identity"))
		return
	}

	if err := AddUserToCourseGroup(c, courseID, groupName, targetUserID, callerUserID); err != nil {
		writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// removeCourseStaffMember godoc
// @Summary Remove a Keycloak user from the Lecturer or Editor group of a course
// @Tags keycloak
// @Produce json
// @Param courseID path string true "Course UUID"
// @Param groupName path string true "Group name (Lecturer or Editor)"
// @Param userID path string true "Keycloak user ID"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /keycloak/{courseID}/group/{groupName}/members/{userID} [delete]
func removeCourseStaffMember(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("courseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	groupName := c.Param("groupName")
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		handleError(c, http.StatusBadRequest, errors.New("userID is required"))
		return
	}

	callerUserID := c.GetString(keycloakTokenVerifier.CtxUserID)
	if callerUserID == "" {
		// Defence in depth: the auth middleware should always set the caller's
		// Keycloak `sub` here. If it didn't, refuse the mutation - an empty
		// callerUserID would also defeat the self-removal guard.
		handleError(c, http.StatusUnauthorized, errors.New("missing caller identity"))
		return
	}

	if err := RemoveUserFromCourseGroup(c, courseID, groupName, targetUserID, callerUserID); err != nil {
		writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// searchKeycloakUsers godoc
// @Summary Search Keycloak users by name, email, or username
// @Tags keycloak
// @Produce json
// @Param q query string true "Search query (>=2 characters)"
// @Param limit query int false "Max results (default 20, max 50)"
// @Success 200 {object} keycloakRealmDTO.UserSearchResults
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /keycloak/users/search [get]
func searchKeycloakUsers(c *gin.Context) {
	query := c.Query("q")

	limit := 20
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			handleError(c, http.StatusBadRequest, errors.New("limit must be an integer"))
			return
		}
		limit = parsed
	}

	results, err := SearchKeycloakUsers(c, query, limit)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, results)
}

// getKeycloakStatus godoc
// @Summary Probe the Keycloak service-account configuration
// @Tags keycloak
// @Produce json
// @Success 200 {object} keycloakRealmDTO.KeycloakStatus
// @Router /keycloak/status [get]
func getKeycloakStatus(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, GetKeycloakStatus(c))
}

// writeServiceError maps service-layer sentinel errors to HTTP status codes.
func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidGroupName), errors.Is(err, ErrSelfRemoval), errors.Is(err, ErrInvalidQuery):
		handleError(c, http.StatusBadRequest, err)
	case errors.Is(err, ErrUserNotFound):
		handleError(c, http.StatusNotFound, err)
	default:
		handleError(c, http.StatusInternalServerError, err)
	}
}
