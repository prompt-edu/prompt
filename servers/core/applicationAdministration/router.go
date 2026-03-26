package applicationAdministration

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration/applicationDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation/coursePhaseParticipationDTO"
	"github.com/prompt-edu/prompt/servers/core/mailing"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// setupApplicationRouter sets up the application endpoints
// @Summary Application Endpoints
// @Description Endpoints for managing applications and application forms
// @Tags applications
// @Security BearerAuth
func setupApplicationRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, applicationMiddleware func() gin.HandlerFunc, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	application := router.Group("/applications", authMiddleware())

	// Application Form Endpoints
	application.GET("/:coursePhaseID/form", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), getApplicationForm)
	application.PUT("/:coursePhaseID/form", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), updateApplicationForm)
	application.GET("/:coursePhaseID/score", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), getAdditionalScores)
	application.POST("/:coursePhaseID/score", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), uploadAdditionalScore)
	application.PUT("/:coursePhaseID/assessment", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), updateApplicationsStatus)

	application.POST("/:coursePhaseID", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), postApplicationManual)
	application.DELETE("/:coursePhaseID", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), deleteApplications)
	application.GET("/:coursePhaseID/files/:fileId/download-url", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), getApplicationFileDownloadURL)

	application.GET("/:coursePhaseID/:courseParticipationID", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), getApplicationByCPID)
	application.PUT("/:coursePhaseID/:courseParticipationID/assessment", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), updateApplicationAssessment)

	application.GET("/:coursePhaseID/participations", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), getAllApplicationParticipations)

	// Apply Endpoints - No Authentication needed
	apply := router.Group("/apply")
	apply.GET("", getAllOpenApplications)
	apply.GET("/:coursePhaseID", getApplicationFormWithCourseDetails)
	apply.POST("/:coursePhaseID", postApplicationExtern)
	apply.POST("/:coursePhaseID/files/presign", presignApplicationUploadExternal)
	apply.POST("/:coursePhaseID/files/complete", completeApplicationUploadExternal)

	applyAuthenticated := router.Group("/apply/authenticated", applicationMiddleware())
	applyAuthenticated.GET("/:coursePhaseID", getApplicationAuthenticated)
	applyAuthenticated.POST("/:coursePhaseID", postApplicationAuthenticated)
	applyAuthenticated.POST("/:coursePhaseID/files/presign", presignApplicationUploadAuthenticated)
	applyAuthenticated.POST("/:coursePhaseID/files/complete", completeApplicationUploadAuthenticated)
	applyAuthenticated.DELETE("/:coursePhaseID/files/:fileId", deleteApplicationFileAuthenticated)

}

// getApplicationForm godoc
// @Summary Get application form
// @Description Get the application form for a course phase
// @Tags applications
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} applicationDTO.Form
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID}/form [get]
func getApplicationForm(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	applicationForm, err := GetApplicationForm(c, coursePhaseId)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not get application form"))
		return
	}

	c.IndentedJSON(http.StatusOK, applicationForm)

}

// updateApplicationForm godoc
// @Summary Update application form
// @Description Update the application form for a course phase
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param updatedApplicationForm body applicationDTO.UpdateForm true "Updated application form"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID}/form [put]
func updateApplicationForm(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var updatedApplicationForm applicationDTO.UpdateForm
	if err := c.BindJSON(&updatedApplicationForm); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = validateUpdateForm(c, coursePhaseId, updatedApplicationForm)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = UpdateApplicationForm(c, coursePhaseId, updatedApplicationForm)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not update application form"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application form updated"})
}

// getAllOpenApplications godoc
// @Summary Get all open applications
// @Description Get all open application phases
// @Tags applications
// @Produce json
// @Success 200 {array} applicationDTO.OpenApplication
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply [get]
func getAllOpenApplications(c *gin.Context) {
	openApplications, err := GetOpenApplicationPhases(c)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not get open applications"))
		return
	}

	c.IndentedJSON(http.StatusOK, openApplications)
}

// getApplicationFormWithCourseDetails godoc
// @Summary Get application form with course details
// @Description Get the application form and course details for a course phase
// @Tags applications
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} applicationDTO.FormWithDetails
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply/{coursePhaseID} [get]
func getApplicationFormWithCourseDetails(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	applicationForm, err := GetApplicationFormWithDetails(c, coursePhaseId)
	if err != nil {
		log.Error(err)
		if errors.Is(err, ErrNotFound) {
			handleError(c, http.StatusNotFound, errors.New("application not found"))
			return
		}
		handleError(c, http.StatusInternalServerError, errors.New("could not get application form"))
		return
	}

	c.IndentedJSON(http.StatusOK, applicationForm)
}

// getApplicationAuthenticated godoc
// @Summary Get authenticated application form
// @Description Get the application form for an authenticated user
// @Tags applications
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} applicationDTO.Application
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply/authenticated/{coursePhaseID} [get]
func getApplicationAuthenticated(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	matriculationNumber := c.GetString("matriculationNumber")

	universityLogin := c.GetString("universityLogin")
	if universityLogin == "" {
		handleError(c, http.StatusUnauthorized, errors.New("no university login found"))
		return
	}

	firstName := c.GetString("firstName")
	lastName := c.GetString("lastName")

	applicationForm, err := GetApplicationAuthenticatedByMatriculationNumberAndUniversityLogin(c, coursePhaseID, matriculationNumber, universityLogin)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not get application form"))
		return
	}

	if applicationForm.Student != nil && firstName != "" && lastName != "" {
		applicationForm.Student.FirstName = firstName
		applicationForm.Student.LastName = lastName
	}

	c.IndentedJSON(http.StatusOK, applicationForm)

}

// postApplicationManual godoc
// @Summary Manually post an application
// @Description Post an application for a student (manual, authenticated)
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param application body applicationDTO.PostApplication true "Application to post"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 405 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID} [post]
func postApplicationManual(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	var application applicationDTO.PostApplication
	if err := c.BindJSON(&application); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = validateApplicationManualAdd(c, coursePhaseId, application)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := PostApplicationAuthenticatedStudent(c, coursePhaseId, application)
	if err != nil {
		log.Error(err)
		if errors.Is(err, ErrAlreadyApplied) {
			handleError(c, http.StatusMethodNotAllowed, errors.New("already applied"))
			return
		}

		handleError(c, http.StatusInternalServerError, errors.New("could not post application"))
		return
	}

	confirmationMailSent, err := mailing.SendApplicationConfirmationMail(c, coursePhaseId, courseParticipationID)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not send confirmation mail"))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "application posted", "confirmationMailSent": confirmationMailSent})
}

// postApplicationExtern godoc
// @Summary Post an application (external)
// @Description Post an application for a student (external, unauthenticated)
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param application body applicationDTO.PostApplication true "Application to post"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 405 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply/{coursePhaseID} [post]
func postApplicationExtern(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var application applicationDTO.PostApplication
	if err := c.BindJSON(&application); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = validateApplication(c, coursePhaseId, application, false)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := PostApplicationExtern(c, coursePhaseId, application)
	if err != nil {
		log.Error(err)
		if errors.Is(err, ErrAlreadyApplied) {
			handleError(c, http.StatusMethodNotAllowed, errors.New("already applied"))
			return
		} else if errors.Is(err, ErrStudentDetailsDoNotMatch) {
			handleError(c, http.StatusConflict, errors.New("student exists but details do not match"))
			return
		}

		handleError(c, http.StatusInternalServerError, errors.New("could not post application"))
		return
	}

	confirmationMailSent, err := mailing.SendApplicationConfirmationMail(c, coursePhaseId, courseParticipationID)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not send confirmation mail"))
		return
	}

	// return if mailing is configured.
	c.JSON(http.StatusCreated, gin.H{"message": "application posted", "confirmationMailSent": confirmationMailSent})
}

// postApplicationAuthenticated godoc
// @Summary Post an application (authenticated)
// @Description Post an application for a student (authenticated)
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param application body applicationDTO.PostApplication true "Application to post"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply/authenticated/{coursePhaseID} [post]
func postApplicationAuthenticated(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	userEmail := c.GetString("userEmail")
	matriculationNumber := c.GetString("matriculationNumber")
	universityLogin := c.GetString("universityLogin")
	firstName := c.GetString("firstName")
	lastName := c.GetString("lastName")
	if userEmail == "" {
		handleError(c, http.StatusUnauthorized, errors.New("no user email found"))
		return
	}

	var application applicationDTO.PostApplication
	if err := c.BindJSON(&application); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = validateApplication(c, coursePhaseId, application, true)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if application.Student.Email != userEmail ||
		application.Student.MatriculationNumber != matriculationNumber ||
		application.Student.UniversityLogin != universityLogin {
		handleError(c, http.StatusUnauthorized, errors.New("credentials do not match payload"))
		return
	}

	if firstName != "" && lastName != "" {
		application.Student.FirstName = firstName
		application.Student.LastName = lastName
	}

	courseParticipationID, err := PostApplicationAuthenticatedStudent(c, coursePhaseId, application)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not post application"))
		return
	}

	confirmationMailSent, err := mailing.SendApplicationConfirmationMail(c, coursePhaseId, courseParticipationID)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not send confirmation mail"))
		return
	}

	// return if mailing is configured.
	c.JSON(http.StatusCreated, gin.H{"message": "application posted", "confirmationMailSent": confirmationMailSent})
}

// getApplicationByCPID godoc
// @Summary Get application by course phase and participation ID
// @Description Get an application by course phase ID and course participation ID
// @Tags applications
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Success 200 {object} applicationDTO.Application
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID}/{courseParticipationID} [get]
func getApplicationByCPID(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	application, err := GetApplicationByCPID(c, coursePhaseId, courseParticipationID)
	if err != nil {
		log.Error(err)
		if errors.Is(err, ErrNotFound) {
			handleError(c, http.StatusNotFound, errors.New("application not found"))
			return
		}
		handleError(c, http.StatusInternalServerError, errors.New("could not get application"))
		return
	}

	c.IndentedJSON(http.StatusOK, application)
}

// getAllApplicationParticipations godoc
// @Summary Get all application participations
// @Description Get all participations for a course phase
// @Tags applications
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} applicationDTO.ApplicationParticipation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID}/participations [get]
func getAllApplicationParticipations(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	applications, err := GetAllApplicationParticipations(c, coursePhaseId)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not get applications"))
		return
	}

	c.IndentedJSON(http.StatusOK, applications)
}

// updateApplicationAssessment godoc
// @Summary Update application assessment
// @Description Update the assessment for an application
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Param assessment body applicationDTO.PutAssessment true "Assessment to update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID}/{courseParticipationID}/assessment [put]
func updateApplicationAssessment(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationId, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var assessment applicationDTO.PutAssessment
	if err := c.BindJSON(&assessment); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = validateUpdateAssessment(c, coursePhaseId, courseParticipationId, assessment)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = UpdateApplicationAssessment(c, coursePhaseId, courseParticipationId, assessment)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not update application assessment"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application assessment updated"})
}

// uploadAdditionalScore godoc
// @Summary Upload additional score
// @Description Upload an additional score for a course phase
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param additionalScore body applicationDTO.AdditionalScoreUpload true "Additional score to upload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID}/score [post]
func uploadAdditionalScore(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var additionalScore applicationDTO.AdditionalScoreUpload
	if err := c.BindJSON(&additionalScore); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = validateAdditionalScore(additionalScore)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = UploadAdditionalScore(c, coursePhaseId, additionalScore)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not upload additional score"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "additional score uploaded"})
}

// getAdditionalScores godoc
// @Summary Get additional scores
// @Description Get additional scores for a course phase
// @Tags applications
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} applicationDTO.AdditionalScore
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID}/score [get]
func getAdditionalScores(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	additionalScore, err := GetAdditionalScores(c, coursePhaseId)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not get additional score"))
		return
	}

	c.IndentedJSON(http.StatusOK, additionalScore)
}

// updateApplicationsStatus godoc
// @Summary Update applications status
// @Description Batch update the status of multiple applications
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param status body coursePhaseParticipationDTO.UpdateCoursePhaseParticipationStatus true "Status update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID}/assessment [put]
func updateApplicationsStatus(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var status coursePhaseParticipationDTO.UpdateCoursePhaseParticipationStatus
	if err := c.BindJSON(&status); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	participationIDs, err := coursePhaseParticipation.BatchUpdatePassStatus(c, coursePhaseId, status.CourseParticipationIDs, status.PassStatus)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not update application status"))
		return
	}
	log.Info("Updated ", len(participationIDs), " participations")

	c.JSON(http.StatusOK, gin.H{"message": "application status updated"})
}

// deleteApplications godoc
// @Summary Delete applications
// @Description Delete applications for a course phase
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationIDs body []string true "Course Participation UUIDs to delete"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID} [delete]
func deleteApplications(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var courseParticipationIDs []uuid.UUID
	if err := c.BindJSON(&courseParticipationIDs); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = DeleteApplications(c, coursePhaseId, courseParticipationIDs)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("could not delete applications"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "applications deleted"})
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
