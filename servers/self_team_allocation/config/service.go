package config

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
)

type ConfigService struct {
	queries db.Queries
}

func NewConfigService(queries db.Queries) *ConfigService {
	return &ConfigService{
		queries: queries,
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
	surveyTimeframe, err := s.queries.GetTimeframe(ctx, coursePhaseID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return map[string]bool{
			"surveyTimeframe": false,
		}, nil
	} else if err != nil {
		return nil, err
	}

	timeframeSet := surveyTimeframe.Starttime.Valid && surveyTimeframe.Endtime.Valid

	return map[string]bool{
		"surveyTimeframe": timeframeSet,
	}, nil
}
