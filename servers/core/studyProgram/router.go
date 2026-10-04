package studyProgram

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/studyProgram/studyProgramDTO"
	log "github.com/sirupsen/logrus"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *StudyProgramService, authMiddleware func() gin.HandlerFunc) {
	setupStudyProgramRouter(routerGroup, service, authMiddleware, permissionValidation.CheckAccessControlByRole)
}

func setupStudyProgramRouter(router *gin.RouterGroup, s *StudyProgramService, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	studyProgramRouter := router.Group("/study-programs")
	studyProgramRouter.GET("", s.getAllStudyPrograms)

	adminRouter := studyProgramRouter.Group("", authMiddleware(), permissionRoleMiddleware(permissionValidation.PromptAdmin))
	adminRouter.GET("/student-counts", s.getStudentCounts)
	adminRouter.POST("", s.createStudyProgram)
	adminRouter.PUT("/:study-program-uuid", s.updateStudyProgram)
	adminRouter.DELETE("/:study-program-uuid", s.deleteStudyProgram)
}

// getAllStudyPrograms godoc
// @Summary Get all study programs
// @Description Get the study programs applicants can pick, ordered by name. No authentication required.
// @Tags studyPrograms
// @Produce json
// @Success 200 {object} []studyProgramDTO.StudyProgram
// @Failure 500 {object} utils.ErrorResponse
// @Router /study-programs [get]
func (s *StudyProgramService) getAllStudyPrograms(c *gin.Context) {
	studyPrograms, err := s.ListStudyPrograms(c)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to get study programs"))
		return
	}
	c.IndentedJSON(http.StatusOK, studyPrograms)
}

// getStudentCounts godoc
// @Summary Get the number of students per study program
// @Description Count the students whose stored study program matches each listed program
// @Tags studyPrograms
// @Produce json
// @Success 200 {object} []studyProgramDTO.StudyProgramStudentCount
// @Failure 500 {object} utils.ErrorResponse
// @Router /study-programs/student-counts [get]
func (s *StudyProgramService) getStudentCounts(c *gin.Context) {
	counts, err := s.GetStudentCounts(c)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to count students per study program"))
		return
	}
	c.IndentedJSON(http.StatusOK, counts)
}

// createStudyProgram godoc
// @Summary Create a study program
// @Description Add a study program to the list applicants can pick
// @Tags studyPrograms
// @Accept json
// @Produce json
// @Param studyProgram body studyProgramDTO.CreateStudyProgram true "Study program to create"
// @Success 201 {object} studyProgramDTO.StudyProgram
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /study-programs [post]
func (s *StudyProgramService) createStudyProgram(c *gin.Context) {
	var input studyProgramDTO.CreateStudyProgram
	if err := c.BindJSON(&input); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := validateStudyProgram(input.Name, input.ShortName); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	studyProgram, err := s.CreateStudyProgram(c, input)
	if err != nil {
		handleServiceError(c, err, "failed to create study program")
		return
	}
	c.IndentedJSON(http.StatusCreated, studyProgram)
}

// updateStudyProgram godoc
// @Summary Update a study program
// @Description Rename a study program or change its short name. A rename also updates every student with the previous name.
// @Tags studyPrograms
// @Accept json
// @Produce json
// @Param study-program-uuid path string true "Study program UUID"
// @Param studyProgram body studyProgramDTO.UpdateStudyProgram true "Updated study program"
// @Success 200 {object} studyProgramDTO.StudyProgram
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /study-programs/{study-program-uuid} [put]
func (s *StudyProgramService) updateStudyProgram(c *gin.Context) {
	id, err := uuid.Parse(c.Param("study-program-uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var input studyProgramDTO.UpdateStudyProgram
	if err := c.BindJSON(&input); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := validateStudyProgram(input.Name, input.ShortName); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	studyProgram, err := s.UpdateStudyProgram(c, id, input)
	if err != nil {
		handleServiceError(c, err, "failed to update study program")
		return
	}
	c.IndentedJSON(http.StatusOK, studyProgram)
}

// deleteStudyProgram godoc
// @Summary Delete a study program
// @Description Remove a study program from the list. Students keep their stored study program.
// @Tags studyPrograms
// @Param study-program-uuid path string true "Study program UUID"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /study-programs/{study-program-uuid} [delete]
func (s *StudyProgramService) deleteStudyProgram(c *gin.Context) {
	id, err := uuid.Parse(c.Param("study-program-uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := s.DeleteStudyProgram(c, id); err != nil {
		handleServiceError(c, err, "failed to delete study program")
		return
	}
	c.Status(http.StatusNoContent)
}

func handleServiceError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, ErrStudyProgramNotFound):
		handleError(c, http.StatusNotFound, err)
	case errors.Is(err, ErrDuplicateStudyProgram):
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
