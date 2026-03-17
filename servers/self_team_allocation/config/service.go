package config

import (
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
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
	surveyTimeframe, err := ConfigServiceSingleton.queries.GetTimeframe(ctx, coursePhaseID)
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
