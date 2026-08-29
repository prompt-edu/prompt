package config

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/survey/surveyDTO"
)

type surveyTimeframeProvider interface {
	GetSurveyTimeframe(ctx context.Context, coursePhaseID uuid.UUID) (surveyDTO.SurveyTimeframe, error)
}

type ConfigService struct {
	queries db.Queries
	survey  surveyTimeframeProvider
}

func NewConfigService(queries db.Queries, survey surveyTimeframeProvider) *ConfigService {
	return &ConfigService{
		queries: queries,
		survey:  survey,
	}
}

type configHandler struct {
	service *ConfigService
}

func (h *configHandler) HandlePhaseConfig(c *gin.Context) (map[string]bool, error) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		return nil, err
	}

	return h.service.GetPhaseConfig(c.Request.Context(), coursePhaseID)
}

func (s *ConfigService) GetPhaseConfig(ctx context.Context, coursePhaseID uuid.UUID) (map[string]bool, error) {
	surveyTimeframe, err := s.survey.GetSurveyTimeframe(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}

	teams, err := s.queries.GetTeamsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	teamsExist := len(teams) > 0

	skills, err := s.queries.GetSkillsByCoursePhase(ctx, coursePhaseID)
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
