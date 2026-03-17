package teams

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/team/teamDTO"
	log "github.com/sirupsen/logrus"
)

func setupTeamRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	teamRouter := routerGroup.Group("/team")

	teamRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseStudent), getAllTeams)
	teamRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseStudent), createTeams)
	teamRouter.PUT("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), updateTeam)
	// only allowing student - as this is a self assignment
	teamRouter.PUT("/:teamID/assignment", authMiddleware(promptSDK.CourseStudent), assignTeam)
	teamRouter.DELETE("/:teamID/assignment", authMiddleware(promptSDK.CourseStudent), leaveTeam)
	teamRouter.DELETE("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), deleteTeam)

	// this is required to comply with the inter phase communication protocol
	teamRouter.GET("/:teamID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTeamByID)

	// Tutor management endpoints
	teamRouter.POST("/tutors", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), importTutors)
	teamRouter.POST("/:teamID/tutor", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), createManualTutor)
	teamRouter.DELETE("/tutor/:tutorID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), deleteTutor)
	teamRouter.GET("/tutors", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTutors)
}

// getAllTeams godoc
// @Summary Get all teams
// @Description Get all teams for a course phase
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} promptTypes.Team
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team [get]
func getAllTeams(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	teams, err := GetAllTeams(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, teams)
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
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/{teamID} [get]
func getTeamByID(c *gin.Context) {
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

	team, err := GetTeamByID(c, coursePhaseID, teamID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
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
func createTeams(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	allowed, err := ValidateTimeframe(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if !allowed {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request teamDTO.CreateTeamsRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = CreateNewTeams(c, request.TeamNames, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
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
func updateTeam(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	allowed, err := ValidateTimeframe(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if !allowed {
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

	err = UpdateTeam(c, coursePhaseID, teamID, request.NewTeamName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
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
func assignTeam(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	allowed, err := ValidateTimeframe(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if !allowed {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	teamID, err := uuid.Parse(c.Param("teamID"))
	if err != nil {
		log.Error("Error parsing teamID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// Check if the team has already 3 members
	team, err := GetTeamByID(c, coursePhaseID, teamID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	// TODO remove magic number here
	if len(team.Members) >= 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team is already full"})
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

	err = AssignTeam(c, coursePhaseID, teamID, courseParticipationID.(uuid.UUID), firstName.(string), lastName.(string))
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
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
func leaveTeam(c *gin.Context) {
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

	courseParticipationID, ok := c.Get("courseParticipationID")
	if !ok {
		log.Error("Error getting courseParticipationID from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "courseParticipationID not found"})
		return
	}

	allowed, err := ValidateTimeframe(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if !allowed {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = LeaveTeam(c, coursePhaseID, teamID, courseParticipationID.(uuid.UUID))
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
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
func deleteTeam(c *gin.Context) {
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

	err = DeleteTeam(c, coursePhaseID, teamID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

// importTutors godoc
// @Summary Import tutors
// @Description Import multiple tutors for the course phase
// @Tags team
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param tutors body []promptTypes.Person true "List of tutors to import"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/tutors [post]
func importTutors(c *gin.Context) {
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

	if err := ImportTutors(c, coursePhaseID, tutors); err != nil {
		log.Error("Error importing tutors: ", err)
		handleError(c, http.StatusInternalServerError, err)
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
func createManualTutor(c *gin.Context) {
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

	var request struct {
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
	}

	if err := c.BindJSON(&request); err != nil {
		log.Error("Error binding request: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := CreateManualTutor(c, coursePhaseID, request.FirstName, request.LastName, teamID); err != nil {
		log.Error("Error creating manual tutor: ", err)
		handleError(c, http.StatusInternalServerError, err)
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
func deleteTutor(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	tutorID, err := uuid.Parse(c.Param("tutorID"))
	if err != nil {
		log.Error("Error parsing tutorID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := DeleteTutor(c, coursePhaseID, tutorID); err != nil {
		log.Error("Error deleting tutor: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// getTutors godoc
// @Summary Get all tutors
// @Description Get all tutors for a course phase
// @Tags team
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} promptTypes.Person
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/team/tutors [get]
func getTutors(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	tutors, err := GetTutorsByCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, tutors)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
