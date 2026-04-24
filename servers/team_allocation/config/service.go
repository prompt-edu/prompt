package config

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/survey"
)

type ConfigService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var ConfigServiceSingleton *ConfigService

type ConfigHandler struct{}

func (h *ConfigHandler) HandlePhaseConfig(c *gin.Context) (config map[string]bool, err error) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		return nil, err
	}

	ctx := c.Request.Context()
	surveyTimeframe, err := survey.GetSurveyTimeframe(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}

	teams, err := survey.GetActiveSurveyTeams(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	teamsExist := len(teams) > 0

	skills, err := ConfigServiceSingleton.queries.GetSkillsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	skillsExist := len(skills) > 0

	return map[string]bool{
		"surveyTimeframe": surveyTimeframe.TimeframeSet,
		"teams":           teamsExist,
		"skills":          skillsExist,
	}, nil
}
