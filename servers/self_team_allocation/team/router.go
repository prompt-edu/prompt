package teams

import (
	"errors"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/tutorscope"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/team/teamDTO"
	log "github.com/sirupsen/logrus"
)

type teamHandler struct {
	teamsService      *TeamsService
	assignmentService *AssignmentService
}

func RegisterRoutes(routerGroup *gin.RouterGroup, teamsService *TeamsService, assignmentService *AssignmentService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	handler := &teamHandler{teamsService: teamsService, assignmentService: assignmentService}
	teamRouter := routerGroup.Group("/team")
	scopingMW := TutorScopeMiddleware(teamsService.conn)

	teamRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, handler.getAllTeams)
	teamRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseStudent), handler.createTeams)
	teamRouter.PUT("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), handler.updateTeam)
	// only allowing student - as this is a self assignment
	teamRouter.PUT("/:teamID/assignment", audit.Describe("Joined a team"), authMiddleware(promptSDK.CourseStudent), handler.assignTeam)
	teamRouter.DELETE("/:teamID/assignment", audit.Describe("Left a team"), authMiddleware(promptSDK.CourseStudent), handler.leaveTeam)
	teamRouter.DELETE("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), handler.deleteTeam)

	// this is required to comply with the inter phase communication protocol
	teamRouter.GET("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), scopingMW, handler.getTeamByID)

	// Tutor management endpoints
	teamRouter.POST("/tutors", audit.Describe("Imported tutors"), authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), handler.importTutors)
	teamRouter.POST("/:teamID/tutor", audit.Describe("Added a tutor to a team"), authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), handler.createManualTutor)
	teamRouter.DELETE("/tutor/:tutorID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), handler.deleteTutor)
	// Lecturers only, since it carries the tutors' logins: GET /team also serves students and tutors.
	teamRouter.GET("/tutors", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), handler.getTutorAssignments)
	teamRouter.GET("/tutors/me", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), handler.getMyTutor)
}

// checkTimeframe answers the request itself and returns false when team changes are not
// allowed right now: 400 outside the configured window, 500 if the window cannot be loaded.
func (h *teamHandler) checkTimeframe(c *gin.Context, coursePhaseID uuid.UUID) bool {
	allowed, err := h.teamsService.ValidateTimeframe(c, coursePhaseID)
	if errors.Is(err, ErrOutsideTimeframe) || (err == nil && !allowed) {
		sdkUtils.HandleError(c, http.StatusBadRequest, ErrOutsideTimeframe)
		return false
	}
	if err != nil {
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return false
	}
	return true
}

// teamCreatorID returns the creating student's course participation, or uuid.Nil for admins
// and lecturers, whose team creations are not limited. ok is false for a student without one.
func teamCreatorID(c *gin.Context) (uuid.UUID, bool) {
	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if ok && (tokenUser.Roles[promptSDK.PromptAdmin] || tokenUser.IsLecturer) {
		return uuid.Nil, true
	}
	courseParticipationID, err := keycloakTokenVerifier.GetUserCourseParticipationID(c)
	if err != nil || courseParticipationID == uuid.Nil {
		log.Error("Error getting courseParticipationID from context: ", err)
		return uuid.Nil, false
	}
	return courseParticipationID, true
}

func teamErrorStatus(err error) int {
	switch {
	case errors.Is(err, ErrInvalidTeamName):
		return http.StatusBadRequest
	case errors.Is(err, ErrDuplicateTeamName), errors.Is(err, ErrTeamLimitReached):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
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
func (h *teamHandler) getAllTeams(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	teams, err := h.teamsService.GetAllTeams(c, coursePhaseID)
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
func (h *teamHandler) getTeamByID(c *gin.Context) {
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

	team, err := h.teamsService.GetTeamByID(c, coursePhaseID, teamID)
	if err != nil {
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
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
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team [post]
func (h *teamHandler) createTeams(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if !h.checkTimeframe(c, coursePhaseID) {
		return
	}

	var request teamDTO.CreateTeamsRequest
	if err := c.BindJSON(&request); err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	creatorID, ok := teamCreatorID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "courseParticipationID not found"})
		return
	}

	err = h.teamsService.CreateNewTeams(c, request.TeamNames, coursePhaseID, creatorID)
	if err != nil {
		sdkUtils.HandleError(c, teamErrorStatus(err), err)
		return
	}
	c.Status(http.StatusCreated)
}

// updateTeam godoc
// @Summary Update team
// @Description Update team name
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
func (h *teamHandler) updateTeam(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if !h.checkTimeframe(c, coursePhaseID) {
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

	err = h.teamsService.UpdateTeam(c, coursePhaseID, teamID, request.NewTeamName)
	if err != nil {
		sdkUtils.HandleError(c, teamErrorStatus(err), err)
		return
	}
	c.Status(http.StatusOK)
}

// assignTeam godoc
// @Summary Assign student to team
// @Description Assign the authenticated student to a team
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param teamID path string true "Team UUID"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/{teamID}/assignment [put]
func (h *teamHandler) assignTeam(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if !h.checkTimeframe(c, coursePhaseID) {
		return
	}

	teamID, err := uuid.Parse(c.Param("teamID"))
	if err != nil {
		log.Error("Error parsing teamID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, ok := c.Get("courseParticipationID")
	if !ok {
		log.Error("Error getting courseParticipationID from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "courseParticipationID not found"})
		return
	}

	firstName, ok := c.Get("firstName")
	if !ok {
		log.Error("Error getting student name from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "student name not found"})
		return
	}

	lastName, ok := c.Get("lastName")
	if !ok {
		log.Error("Error getting student name from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "student name not found"})
		return
	}

	err = h.assignmentService.AssignTeam(c, coursePhaseID, teamID, courseParticipationID.(uuid.UUID), firstName.(string), lastName.(string))
	if err != nil {
		sdkUtils.HandleError(c, assignTeamErrorStatus(err), err)
		return
	}
	c.Status(http.StatusOK)
}

// leaveTeam godoc
// @Summary Leave team
// @Description Remove the authenticated student from a team
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param teamID path string true "Team UUID"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/{teamID}/assignment [delete]
func (h *teamHandler) leaveTeam(c *gin.Context) {
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

	courseParticipationID, ok := c.Get("courseParticipationID")
	if !ok {
		log.Error("Error getting courseParticipationID from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "courseParticipationID not found"})
		return
	}

	if !h.checkTimeframe(c, coursePhaseID) {
		return
	}

	err = h.assignmentService.LeaveTeam(c, coursePhaseID, teamID, courseParticipationID.(uuid.UUID))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

// deleteTeam godoc
// @Summary Delete team
// @Description Delete a team from the course phase
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param teamID path string true "Team UUID"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/{teamID} [delete]
func (h *teamHandler) deleteTeam(c *gin.Context) {
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

	err = h.teamsService.DeleteTeam(c, coursePhaseID, teamID)
	if err != nil {
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
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
func (h *teamHandler) importTutors(c *gin.Context) {
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

	if err := h.teamsService.ImportTutors(c, coursePhaseID, tutors); err != nil {
		if errors.Is(err, ErrMissingLogin) || errors.Is(err, ErrMissingTutorID) || errors.Is(err, ErrDuplicateLogin) ||
			errors.Is(err, ErrDuplicateTutorID) || errors.Is(err, ErrNoTeams) || errors.Is(err, ErrInvalidTeamForPhase) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Error("Error importing tutors: ", err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// createManualTutor godoc
// @Summary Create manual tutor
// @Description Create a manual tutor for a specific team
// @Tags team
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param teamID path string true "Team UUID"
// @Param request body object{firstName=string,lastName=string} true "Tutor information"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/{teamID}/tutor [post]
func (h *teamHandler) createManualTutor(c *gin.Context) {
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

	var request struct {
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
	}

	if err := c.BindJSON(&request); err != nil {
		log.Error("Error binding request: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if err := h.teamsService.CreateManualTutor(c, coursePhaseID, request.FirstName, request.LastName, teamID); err != nil {
		log.Error("Error creating manual tutor: ", err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// deleteTutor godoc
// @Summary Delete tutor
// @Description Delete a tutor from the course phase
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param tutorID path string true "Tutor UUID"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/tutor/{tutorID} [delete]
func (h *teamHandler) deleteTutor(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	tutorID, err := uuid.Parse(c.Param("tutorID"))
	if err != nil {
		log.Error("Error parsing tutorID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if err := h.teamsService.DeleteTutor(c, coursePhaseID, tutorID); err != nil {
		log.Error("Error deleting tutor: ", err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// getTutorAssignments godoc
// @Summary List the teams of every tutor
// @Description One entry per tutor and team, with the id the tutor is stored under and their university login
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} teamDTO.TutorAssignment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/tutors [get]
func (h *teamHandler) getTutorAssignments(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	assignments, err := h.teamsService.GetTutorAssignments(c, coursePhaseID)
	if err != nil {
		log.Error("Error getting tutor assignments: ", err)
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, assignments)
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
func (h *teamHandler) getMyTutor(c *gin.Context) {
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

	tutor, err := h.teamsService.GetTutorByUniversityLogin(c, coursePhaseID, tutorscope.NormalizeLogin(tokenUser.UniversityLogin))
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
