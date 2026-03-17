package infrastructureSetup

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/intro_course/infrastructureSetup/infrastructureDTO"
	log "github.com/sirupsen/logrus"
)

func setupInfrastructureRouter(router *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	infrastructureRouter := router.Group("/infrastructure")

	// Infrastructure setup routes
	infrastructureRouter.POST("/gitlab/course-setup", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), createCourseSetup)
	infrastructureRouter.POST("/gitlab/student-setup/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), setupStudentInfrastructure)

	// Infrastructure status routes
	infrastructureRouter.GET("/gitlab/student-setup", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getAllStudentGitlabStatus)

	// Route for manually overwriting the status (i.e. if instructor manually created or fixed the repo)
	infrastructureRouter.PUT("/gitlab/student-setup/:courseParticipationID/manual", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), manuallyOverwriteStudentGitlabStatus)
}

// createCourseSetup godoc
// @Summary Create course infrastructure
// @Description Creates the GitLab course infrastructure for the course phase.
// @Tags infrastructure
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body infrastructureDTO.CreateCourseInfrastructureRequest true "Course infrastructure request"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/infrastructure/gitlab/course-setup [post]
func createCourseSetup(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// get semester tag (= top level group name)
	var infrastructureRequest infrastructureDTO.CreateCourseInfrastructureRequest
	if err := c.BindJSON(&infrastructureRequest); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// TODO: remove this later - but parts of the infrastructure for ios25 were already done
	semesterTag := strings.ToUpper(infrastructureRequest.SemesterTag)

	err = CreateCourseInfrastructure(coursePhaseID, semesterTag)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// setupStudentInfrastructure godoc
// @Summary Create student infrastructure
// @Description Creates the GitLab repository for a student.
// @Tags infrastructure
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Param request body infrastructureDTO.CreateStudentRepo true "Student repository request"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/infrastructure/gitlab/student-setup/{courseParticipationID} [post]
func setupStudentInfrastructure(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		log.Error("Error parsing courseParticipationID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// get semester tag (= top level group name)
	var infrastructureRequest infrastructureDTO.CreateStudentRepo
	if err := c.BindJSON(&infrastructureRequest); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// TODO: remove this later - but parts of the infrastructure for ios25 were already done
	semesterTag := strings.ToUpper(infrastructureRequest.SemesterTag)

	err = CreateStudentInfrastructure(c, coursePhaseID, courseParticipationID, semesterTag, infrastructureRequest.RepoName, infrastructureRequest.StudentName, infrastructureRequest.SubmissionDeadline)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)

}

// getAllStudentGitlabStatus godoc
// @Summary Get student GitLab status
// @Description Returns GitLab setup status for all course participations.
// @Tags infrastructure
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} infrastructureDTO.GitlabStatus
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/infrastructure/gitlab/student-setup [get]
func getAllStudentGitlabStatus(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	studentInfrastructureStatus, err := GetAllStudentGitlabStatus(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, studentInfrastructureStatus)
}

// manuallyOverwriteStudentGitlabStatus godoc
// @Summary Manually overwrite student GitLab status
// @Description Marks GitLab setup as completed for a student.
// @Tags infrastructure
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/infrastructure/gitlab/student-setup/{courseParticipationID}/manual [put]
func manuallyOverwriteStudentGitlabStatus(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		log.Error("Error parsing courseParticipationID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = ManuallyOverwriteStudentGitlabStatus(c, coursePhaseID, courseParticipationID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully overwritten student gitlab status"})
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
