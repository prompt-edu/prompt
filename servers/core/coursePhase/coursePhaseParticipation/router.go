package coursePhaseParticipation

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation/coursePhaseParticipationDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// setupCoursePhaseParticipationRouter sets up the course phase participation endpoints
// @Summary Course Phase Participation Endpoints
// @Description Endpoints for managing course phase participations
// @Tags course_phase_participation
// @Security BearerAuth
func setupCoursePhaseParticipationRouter(routerGroup *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	courseParticipation := routerGroup.Group("/course_phases/:uuid/participations", authMiddleware())
	courseParticipation.GET("/self", permissionIDMiddleware(permissionValidation.CourseStudent), getOwnCoursePhaseParticipation)
	courseParticipation.GET("", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), getParticipationsForCoursePhase)
	courseParticipation.GET("/:course_participation_id", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), getParticipation)
	courseParticipation.PUT("/:course_participation_id", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), updateCoursePhaseParticipation)
	// allow to modify multiple at once
	courseParticipation.PUT("", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), updateBatchCoursePhaseParticipation)

	// get the students data of the participations
	courseParticipation.GET("/students", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), getStudentsOfCoursePhase)
}

// getOwnCoursePhaseParticipation godoc
// @Summary Get own course phase participation
// @Description Get the participation of the current user in a course phase
// @Tags course_phase_participation
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Success 200 {object} coursePhaseParticipationDTO.CoursePhaseParticipationStudent
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid}/participations/self [get]
func getOwnCoursePhaseParticipation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	matriculationNumber := c.GetString("matriculationNumber")
	universityLogin := c.GetString("universityLogin")

	if matriculationNumber == "" || universityLogin == "" {
		handleError(c, http.StatusUnauthorized, errors.New("missing matriculation number or university login"))
		return
	}

	coursePhaseParticipation, err := GetOwnCoursePhaseParticipation(c, id, matriculationNumber, universityLogin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			handleError(c, http.StatusNotFound, errors.New("course phase participation not found"))
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, coursePhaseParticipation)
}

// getParticipationsForCoursePhase godoc
// @Summary Get all participations for a course phase
// @Description Get all participations for a given course phase
// @Tags course_phase_participation
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Success 200 {object} coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid}/participations [get]
func getParticipationsForCoursePhase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipations, err := GetAllParticipationsForCoursePhase(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courseParticipations)
}

// getParticipation godoc
// @Summary Get a specific participation
// @Description Get a specific participation by course phase and participation ID
// @Tags course_phase_participation
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Param course_participation_id path string true "Course Participation UUID"
// @Success 200 {object} coursePhaseParticipationDTO.GetCoursePhaseParticipation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid}/participations/{course_participation_id} [get]
func getParticipation(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		log.Error("Error parsing course phase ID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := uuid.Parse(c.Param("course_participation_id"))
	if err != nil {
		log.Error("Error parsing course participation ID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipation, err := GetCoursePhaseParticipation(c, coursePhaseID, courseParticipationID)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courseParticipation)
}

// updateCoursePhaseParticipation godoc
// @Summary Update a course phase participation
// @Description Update a specific course phase participation
// @Tags course_phase_participation
// @Accept json
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Param course_participation_id path string true "Course Participation UUID"
// @Param newCourseParticipation body coursePhaseParticipationDTO.CreateCoursePhaseParticipation true "Participation to update"
// @Success 201 {object} coursePhaseParticipationDTO.GetCoursePhaseParticipation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid}/participations/{course_participation_id} [put]
func updateCoursePhaseParticipation(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := uuid.Parse(c.Param("course_participation_id"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var newCourseParticipation coursePhaseParticipationDTO.CreateCoursePhaseParticipation
	if err := c.BindJSON(&newCourseParticipation); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	newCourseParticipation.CoursePhaseID = coursePhaseID
	newCourseParticipation.CourseParticipationID = courseParticipationID

	if err := Validate(newCourseParticipation); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipation, err := CreateOrUpdateCoursePhaseParticipation(c, nil, newCourseParticipation)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, courseParticipation)
}

// updateBatchCoursePhaseParticipation godoc
// @Summary Batch update course phase participations
// @Description Update multiple course phase participations at once
// @Tags course_phase_participation
// @Accept json
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Param updatedCourseParticipationRequest body []coursePhaseParticipationDTO.UpdateCoursePhaseParticipationRequest true "Participations to update"
// @Success 200 {array} uuid.UUID
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid}/participations [put]
func updateBatchCoursePhaseParticipation(c *gin.Context) {
	coursePhaseId, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// we expect an array of updates
	var updatedCourseParticipationRequest []coursePhaseParticipationDTO.UpdateCoursePhaseParticipationRequest
	if err := c.BindJSON(&updatedCourseParticipationRequest); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var createOrUpdateCourseParticipationDTOs []coursePhaseParticipationDTO.CreateCoursePhaseParticipation
	for _, update := range updatedCourseParticipationRequest {
		if update.CoursePhaseID != coursePhaseId {
			handleError(c, http.StatusBadRequest, errors.New("coursePhaseID in request does not match coursePhaseID in URL"))
			return
		}

		dbParticipation := coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
			CourseParticipationID: update.CourseParticipationID,
			CoursePhaseID:         coursePhaseId, // we only update for one coursePhaseID
			PassStatus:            update.PassStatus,
			RestrictedData:        update.RestrictedData,
			StudentReadableData:   update.StudentReadableData,
		}

		// Validate for complete new participations
		if err := Validate(dbParticipation); err != nil {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		createOrUpdateCourseParticipationDTOs = append(createOrUpdateCourseParticipationDTOs, dbParticipation)
	}

	ids, err := UpdateBatchCoursePhaseParticipation(c, createOrUpdateCourseParticipationDTOs)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ids)
}

// getStudentsOfCoursePhase godoc
// @Summary Get students of a course phase
// @Description Get all students of a given course phase
// @Tags course_phase_participation
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Success 200 {array} studentDTO.Student
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid}/participations/students [get]
func getStudentsOfCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	students, err := GetStudentsOfCoursePhase(c, coursePhaseID)
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
