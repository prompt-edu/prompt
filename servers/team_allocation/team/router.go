package teams

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/team_allocation/team/teamDTO"
	"github.com/prompt-edu/prompt/servers/team_allocation/tutorscope"
	log "github.com/sirupsen/logrus"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *TeamsService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	teamRouter := routerGroup.Group("/team")
	scopingMW := promptSDK.TutorScopingMiddleware(tutorscope.NewResolver(service.queries))

	teamRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, service.getAllTeams)
	teamRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.createTeams)
	teamRouter.PUT("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.updateTeam)
	teamRouter.DELETE("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.deleteTeam)

	teamRouter.POST("/student-names", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.addStudentNamesToTeams)

	teamRouter.POST("/tutors", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.importTutors)
	teamRouter.PUT("/tutors/:universityLogin", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.updateTutorTeam)

	// required for inter-phase communication protocol
	teamRouter.GET("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, service.getTeamByID)
}

// getAllTeams godoc
// @Summary Get all teams
// @Description Get all teams for a course phase
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} map[string][]promptTypes.Team
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team [get]
func (s *TeamsService) getAllTeams(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	teams, err := s.GetAllTeams(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	if tutorTeamID, scoped := promptSDK.GetTutorTeamID(c); scoped {
		teams = filterTeamsByID(teams, tutorTeamID)
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func filterTeamsByID(teams []promptTypes.Team, teamID uuid.UUID) []promptTypes.Team {
	for _, t := range teams {
		if t.ID == teamID {
			return []promptTypes.Team{t}
		}
	}
	return []promptTypes.Team{}
}

// getTeamByID godoc
// @Summary Get team by ID
// @Description Get a specific team by its ID
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param teamID path string true "Team UUID"
// @Success 200 {object} promptTypes.Team
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/{teamID} [get]
func (s *TeamsService) getTeamByID(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	teamID, err := uuid.Parse(c.Param("teamID"))
	if err != nil {
		log.Error("Error parsing teamID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if tutorTeamID, scoped := promptSDK.GetTutorTeamID(c); scoped && teamID != tutorTeamID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access restricted to assigned team"})
		return
	}

	team, err := s.GetTeamByID(c, coursePhaseID, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			handleError(c, http.StatusNotFound, err)
		} else {
			handleError(c, http.StatusInternalServerError, err)
		}
		return
	}
	c.JSON(http.StatusOK, team)
}

// createTeams godoc
// @Summary Create new teams
// @Description Create one or more new teams for a course phase
// @Tags team
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body teamDTO.CreateTeamsRequest true "Team names to create"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team [post]
func (s *TeamsService) createTeams(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request teamDTO.CreateTeamsRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.CreateNewTeams(c, request.TeamNames, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusCreated)
}

// updateTeam godoc
// @Summary Update team
// @Description Update a team name
// @Tags team
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param teamID path string true "Team UUID"
// @Param request body teamDTO.UpdateTeamRequest true "New team name"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/{teamID} [put]
func (s *TeamsService) updateTeam(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	teamID, err := uuid.Parse(c.Param("teamID"))
	if err != nil {
		log.Error("Error parsing teamID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request teamDTO.UpdateTeamRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.UpdateTeam(c, coursePhaseID, teamID, request.NewTeamName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

// deleteTeam godoc
// @Summary Delete team
// @Description Delete a team by its ID
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param teamID path string true "Team UUID"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/{teamID} [delete]
func (s *TeamsService) deleteTeam(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	teamID, err := uuid.Parse(c.Param("teamID"))
	if err != nil {
		log.Error("Error parsing teamID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.DeleteTeam(c, coursePhaseID, teamID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}

// addStudentNamesToTeams godoc
// @Summary Add student names to allocations
// @Description Add student first and last names to team allocations
// @Tags team
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body teamDTO.StudentNameUpdateRequest true "Student names per participation ID"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/student-names [post]
func (s *TeamsService) addStudentNamesToTeams(c *gin.Context) {
	var req teamDTO.StudentNameUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := s.AddStudentNamesToAllocations(c, req); err != nil {
		log.Error("Error adding student names to allocations: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// importTutors godoc
// @Summary Import tutors
// @Description Import tutors for teams in a course phase
// @Tags team
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body []teamDTO.Tutor true "Tutors to import"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/tutors [post]
func (s *TeamsService) importTutors(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var tutors []teamDTO.Tutor
	if err := c.BindJSON(&tutors); err != nil {
		log.Error("Error binding tutors: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := s.ImportTutors(c, coursePhaseID, tutors); err != nil {
		if errors.Is(err, ErrDuplicateLogin) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		log.Error("Error importing tutors: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// updateTutorTeam godoc
// @Summary Update tutor team assignment
// @Description Reassign a tutor to a different team by their university login
// @Tags team
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param universityLogin path string true "Tutor university login (trimmed and lowercased; used as a lookup key, not format-validated)"
// @Param request body teamDTO.UpdateTutorTeamRequest true "New team ID"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/tutors/{universityLogin} [put]
func (s *TeamsService) updateTutorTeam(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	universityLogin := teamDTO.NormalizeUniversityLogin(c.Param("universityLogin"))
	if universityLogin == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "university login is required"})
		return
	}

	var req teamDTO.UpdateTutorTeamRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := s.UpdateTutorTeam(c, coursePhaseID, universityLogin, req.TeamID); err != nil {
		if errors.Is(err, ErrTutorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tutor not found"})
			return
		}
		if errors.Is(err, ErrInvalidTeamForPhase) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "team does not belong to this course phase"})
			return
		}
		log.Error("Error updating tutor team: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}
