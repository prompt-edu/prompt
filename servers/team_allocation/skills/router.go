package skills

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/team_allocation/skills/skillDTO"
	log "github.com/sirupsen/logrus"
)

// setupSkillRouter creates a router group for skill endpoints.
func setupSkillRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	skillRouter := routerGroup.Group("/skill")

	skillRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getAllSkills)
	skillRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), createSkills)
	skillRouter.PUT("/:skillID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), updateSkill)
	skillRouter.DELETE("/:skillID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), deleteSkill)
}

// getAllSkills godoc
// @Summary Get all skills
// @Description Get all skills for a course phase
// @Tags skill
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} skillDTO.Skill
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/skill [get]
func getAllSkills(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	skills, err := GetAllSkills(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, skills)
}

// createSkills godoc
// @Summary Create new skills
// @Description Create one or more skills for a course phase
// @Tags skill
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body skillDTO.CreateSkillsRequest true "Skill names to create"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/skill [post]
func createSkills(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request skillDTO.CreateSkillsRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = CreateNewSkills(c, request.SkillNames, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusCreated)
}

// updateSkill godoc
// @Summary Update skill
// @Description Update a skill name
// @Tags skill
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param skillID path string true "Skill UUID"
// @Param request body skillDTO.UpdateSkillRequest true "New skill name"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/skill/{skillID} [put]
func updateSkill(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	skillID, err := uuid.Parse(c.Param("skillID"))
	if err != nil {
		log.Error("Error parsing skillID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request skillDTO.UpdateSkillRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = UpdateSkill(c, coursePhaseID, skillID, request.NewSkillName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

// deleteSkill godoc
// @Summary Delete skill
// @Description Delete a skill by its ID
// @Tags skill
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param skillID path string true "Skill UUID"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/skill/{skillID} [delete]
func deleteSkill(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	skillID, err := uuid.Parse(c.Param("skillID"))
	if err != nil {
		log.Error("Error parsing skillID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = DeleteSkill(c, coursePhaseID, skillID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
