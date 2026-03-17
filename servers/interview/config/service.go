package config

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

type ConfigService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var ConfigServiceSingleton *ConfigService

type ConfigHandler struct{}

func (h *ConfigHandler) HandlePhaseConfig(c *gin.Context) (config map[string]bool, err error) {
	return map[string]bool{}, nil
}
