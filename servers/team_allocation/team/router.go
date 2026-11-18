package teams

import (
	"errors"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/tutorscope"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/team_allocation/team/teamDTO"
	log "github.com/sirupsen/logrus"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *TeamsService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	teamRouter := routerGroup.Group("/team")
	scopingMW := tutorscope.TeamsMiddleware(tutorscope.NewPgxTeamsResolver(service.conn))

	teamRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, service.getAllTeams)
	teamRouter.POST("", audit.Describe("Created teams"), authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.createTeams)
	teamRouter.PUT("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.updateTeam)
	teamRouter.DELETE("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.deleteTeam)

	teamRouter.POST("/student-names", audit.Describe("Added student names to team allocations"), authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.addStudentNamesToTeams)

	// Lecturers only, since it carries the tutors' logins: GET /team also serves students and tutors.
	teamRouter.GET("/tutors", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.getTutorAssignments)
	teamRouter.POST("/tutors", audit.Describe("Imported tutors"), authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.importTutors)
	teamRouter.GET("/tutors/me", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), service.getMyTutor)

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
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team [get]
func (s *TeamsService) getAllTeams(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	teams, err := s.GetAllTeams(c, coursePhaseID)
	if err != nil {
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	if tutorTeamIDs, scoped := tutorscope.TeamIDs(c); scoped {
		teams = filterTeamsByID(teams, tutorTeamIDs)
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func filterTeamsByID(teams []promptTypes.Team, teamIDs []uuid.UUID) []promptTypes.Team {
	result := make([]promptTypes.Team, 0, len(teamIDs))
	for _, t := range teams {
		if slices.Contains(teamIDs, t.ID) {
			result = append(result, t)
		}
	}
	return result
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
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	teamID, err := uuid.Parse(c.Param("teamID"))
	if err != nil {
		log.Error("Error parsing teamID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if tutorTeamIDs, scoped := tutorscope.TeamIDs(c); scoped && !slices.Contains(tutorTeamIDs, teamID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access restricted to assigned teams"})
		return
	}

	team, err := s.GetTeamByID(c, coursePhaseID, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			sdkUtils.HandleError(c, http.StatusNotFound, err)
		} else {
			sdkUtils.HandleError(c, http.StatusInternalServerError, err)
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
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var request teamDTO.CreateTeamsRequest
	if err := c.BindJSON(&request); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.CreateNewTeams(c, request.TeamNames, coursePhaseID)
	if err != nil {
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
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
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	teamID, err := uuid.Parse(c.Param("teamID"))
	if err != nil {
		log.Error("Error parsing teamID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var request teamDTO.UpdateTeamRequest
	if err := c.BindJSON(&request); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.UpdateTeam(c, coursePhaseID, teamID, request.NewTeamName)
	if err != nil {
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
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
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	teamID, err := uuid.Parse(c.Param("teamID"))
	if err != nil {
		log.Error("Error parsing teamID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.DeleteTeam(c, coursePhaseID, teamID)
	if err != nil {
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
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
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var req teamDTO.StudentNameUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if err := s.AddStudentNamesToAllocations(c, coursePhaseID, req); err != nil {
		log.Error("Error adding student names to allocations: ", err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// getTutorAssignments godoc
// @Summary List tutor assignments
// @Description List every team each tutor of the course phase leads, one entry per tutor and team, with the tutor's id and university login (empty for a tutor stored without one).
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} teamDTO.TutorAssignment
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/tutors [get]
func (s *TeamsService) getTutorAssignments(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	assignments, err := s.GetTutorAssignments(c, coursePhaseID)
	if err != nil {
		log.Error("Error getting tutor assignments: ", err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, assignments)
}

// importTutors godoc
// @Summary Import tutors
// @Description Set the teams of each listed tutor. A tutor is identified by their university login, which is required: their previous assignments in the phase are replaced, and an existing tutor keeps their id.
// @Tags team
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body []teamDTO.Tutor true "Tutors to import"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/tutors [post]
func (s *TeamsService) importTutors(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var tutors []teamDTO.Tutor
	if err := c.BindJSON(&tutors); err != nil {
		log.Error("Error binding tutors: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if err := s.ImportTutors(c, coursePhaseID, tutors); err != nil {
		if errors.Is(err, ErrMissingLogin) || errors.Is(err, ErrMissingTutorID) || errors.Is(err, ErrDuplicateLogin) ||
			errors.Is(err, ErrNoTeams) || errors.Is(err, ErrInvalidTeamForPhase) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Error("Error importing tutors: ", err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// getMyTutor godoc
// @Summary Get the caller's tutor record
// @Description Get the tutor that the caller's university login belongs to in this course phase, so other phases can match the caller to a team's tutors
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} promptTypes.Person
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/tutors/me [get]
func (s *TeamsService) getMyTutor(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated user not found in context"})
		return
	}

	tutor, err := s.GetTutorByUniversityLogin(c, coursePhaseID, tutorscope.NormalizeLogin(tokenUser.UniversityLogin))
	if err != nil {
		if errors.Is(err, ErrTutorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "the caller is not a tutor in this course phase"})
			return
		}
		log.Error("Error getting the caller's tutor record: ", err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, tutor)
}
