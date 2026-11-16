package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/allocation"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/bodyLimit"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/config"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/copy"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/privacy"
	teams "github.com/prompt-edu/prompt/servers/self_team_allocation/team"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/timeframe"

	log "github.com/sirupsen/logrus"
)

// @title           PROMPT Self Team Allocation API
// @version         1.0
// @description     This is the self team allocation server of PROMPT.
// @host            localhost:8084
// @BasePath        /self-team-allocation/api
// @externalDocs.description  PROMPT Documentation
// @externalDocs.url          https://prompt-edu.github.io/prompt/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token authentication. Use format: Bearer {token}

func main() {
	err := promptSDK.Bootstrap(promptSDK.ServiceOptions{
		ServiceName:    "self-team-allocation",
		BasePath:       "/self-team-allocation/api",
		DBEnvPrefix:    "SELF_TEAM_ALLOCATION",
		DefaultDBPort:  "5436",
		SentryDSNEnv:   "SENTRY_DSN_SELF_TEAM_ALLOCATION",
		DefaultAddress: "localhost:8084",
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
		},
		RegisterRoutes: registerRoutes,
	})
	if err != nil {
		log.Fatalf("Self team allocation server stopped: %v", err)
	}
}

func registerRoutes(api, coursePhaseApi *gin.RouterGroup, conn *pgxpool.Pool) error {
	// Bootstrap creates coursePhaseApi before calling back, and Gin snapshots a group's chain when
	// a subgroup is created, so the cap goes on both groups. Routes that need more raise it with
	// bodyLimit.WithLimit.
	api.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))
	coursePhaseApi.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))

	query := db.New(conn)

	coursePhaseApi.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello from team self assignment service"})
	})

	timeframeService := timeframe.NewTimeframeService(*query)
	teamsService := teams.NewTeamsService(*query, conn, timeframeService)
	assignmentService := teams.NewAssignmentService(*query, conn)
	allocationService := allocation.NewAllocationService(*query)
	configService := config.NewConfigService(*query)
	privacyService := privacy.NewPrivacyService(*query, conn)

	teams.RegisterRoutes(coursePhaseApi, teamsService, assignmentService, promptSDK.AuthenticationMiddleware)
	timeframe.RegisterRoutes(coursePhaseApi, timeframeService, promptSDK.AuthenticationMiddleware)
	allocation.RegisterRoutes(coursePhaseApi, allocationService, promptSDK.AuthenticationMiddleware)
	copy.RegisterRoutes(api)
	privacy.RegisterRoutes(api, privacyService)

	config.RegisterRoutes(coursePhaseApi, configService)
	return nil
}
