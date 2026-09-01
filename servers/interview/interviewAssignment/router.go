package interviewAssignment

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	interviewAssignmentDTO "github.com/prompt-edu/prompt/servers/interview/interviewAssignment/interviewAssignmentDTO"
	log "github.com/sirupsen/logrus"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *InterviewAssignmentService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	assignmentRouter := routerGroup.Group("/interview-assignments")
	assignmentRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.PromptLecturer, promptSDK.CourseStudent), service.createInterviewAssignment)
	assignmentRouter.POST("/admin", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), service.createInterviewAssignmentAdmin)
	assignmentRouter.GET("/my-assignment", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.PromptLecturer, promptSDK.CourseStudent), service.getMyInterviewAssignment)
	assignmentRouter.DELETE("/:assignmentId", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.PromptLecturer, promptSDK.CourseStudent), service.deleteInterviewAssignment)
}

func writeServiceError(c *gin.Context, err error) {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		c.JSON(serviceErr.StatusCode, gin.H{"error": serviceErr.Message})
		return
	}

	log.Errorf("Unexpected interview assignment service error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}

// createInterviewAssignment godoc
// @Summary Assign current user to an interview slot
// @Description Students can book themselves into an available interview slot
// @Tags interview-assignments
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body interviewAssignmentDTO.CreateInterviewAssignmentRequest true "Assignment details"
// @Success 201 {object} interviewAssignmentDTO.InterviewAssignmentResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-assignments [post]
func (s *InterviewAssignmentService) createInterviewAssignment(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	var req interviewAssignmentDTO.CreateInterviewAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	participationID, err := keycloakTokenVerifier.GetUserCourseParticipationID(c)
	if err != nil {
		c.JSON(keycloakTokenVerifier.GetUserCourseParticipationIDErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	response, err := s.CreateInterviewAssignment(c.Request.Context(), coursePhaseID, participationID, req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// createInterviewAssignmentAdmin godoc
// @Summary Manually assign a student to an interview slot (Admin only)
// @Description Allows admins/lecturers to manually assign any student to an interview slot
// @Tags interview-assignments
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body interviewAssignmentDTO.CreateInterviewAssignmentAdminRequest true "Assignment details with course participation ID"
// @Success 201 {object} interviewAssignmentDTO.InterviewAssignmentResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-assignments/admin [post]
func (s *InterviewAssignmentService) createInterviewAssignmentAdmin(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	var req interviewAssignmentDTO.CreateInterviewAssignmentAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := s.CreateInterviewAssignmentAdmin(c.Request.Context(), coursePhaseID, req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// getMyInterviewAssignment godoc
// @Summary Get current user's interview assignment
// @Description Retrieves the interview assignment for the current student
// @Tags interview-assignments
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} interviewAssignmentDTO.InterviewAssignmentResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-assignments/my-assignment [get]
func (s *InterviewAssignmentService) getMyInterviewAssignment(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	participationID, err := keycloakTokenVerifier.GetUserCourseParticipationID(c)
	if err != nil {
		c.JSON(keycloakTokenVerifier.GetUserCourseParticipationIDErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	response, err := s.GetMyInterviewAssignment(c.Request.Context(), coursePhaseID, participationID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// deleteInterviewAssignment godoc
// @Summary Delete an interview assignment
// @Description Removes a student's assignment to an interview slot
// @Tags interview-assignments
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param assignmentId path string true "Assignment UUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-assignments/{assignmentId} [delete]
func (s *InterviewAssignmentService) deleteInterviewAssignment(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	assignmentID, err := uuid.Parse(c.Param("assignmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has privileged roles
	isPrivileged := tokenUser.Roles[promptSDK.PromptAdmin] ||
		tokenUser.Roles[promptSDK.PromptLecturer] ||
		tokenUser.IsLecturer ||
		tokenUser.IsEditor

	// For non-privileged users, get their participation ID for ownership check
	var selfParticipationID uuid.UUID
	if !isPrivileged {
		selfParticipationID, err = keycloakTokenVerifier.GetUserCourseParticipationID(c)
		if err != nil {
			c.JSON(keycloakTokenVerifier.GetUserCourseParticipationIDErrorStatus(err), gin.H{"error": err.Error()})
			return
		}
	}

	err = s.DeleteInterviewAssignment(c.Request.Context(), coursePhaseID, assignmentID, selfParticipationID, isPrivileged)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
