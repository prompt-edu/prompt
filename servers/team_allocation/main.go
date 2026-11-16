package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation"
	"github.com/prompt-edu/prompt/servers/team_allocation/bodyLimit"
	"github.com/prompt-edu/prompt/servers/team_allocation/config"
	"github.com/prompt-edu/prompt/servers/team_allocation/copy"
	"github.com/prompt-edu/prompt/servers/team_allocation/coursePhaseDeletion"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/privacy"
	"github.com/prompt-edu/prompt/servers/team_allocation/skills"
	"github.com/prompt-edu/prompt/servers/team_allocation/survey"
	teams "github.com/prompt-edu/prompt/servers/team_allocation/team"
	"github.com/prompt-edu/prompt/servers/team_allocation/tease"
	log "github.com/sirupsen/logrus"
)

// @title           PROMPT Team Allocation API
// @version         1.0
// @description     This is the team allocation server of PROMPT.
// @host            localhost:8083
// @BasePath        /team-allocation/api
// @externalDocs.description  PROMPT Documentation
// @externalDocs.url          https://prompt-edu.github.io/prompt/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token authentication. Use format: Bearer {token}

func main() {
	err := promptSDK.Bootstrap(promptSDK.ServiceOptions{
		ServiceName:    "team-allocation",
		BasePath:       "/team-allocation/api",
		DBEnvPrefix:    "TEAM_ALLOCATION",
		DefaultDBPort:  "5434",
		SentryDSNEnv:   "SENTRY_DSN_TEAM_ALLOCATION",
		DefaultAddress: "localhost:8083",
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
			promptTypes.CapabilityPhaseDeletion:   true,
		},
		RegisterRoutes: registerRoutes,
	})
	if err != nil {
		log.Fatalf("Team allocation server stopped: %v", err)
	}
}

func registerRoutes(api, coursePhaseApi *gin.RouterGroup, conn *pgxpool.Pool) error {
	// Bootstrap creates coursePhaseApi before calling back, and Gin snapshots a group's chain when
	// a subgroup is created, so the cap goes on both groups. Routes that need more raise it with
	// bodyLimit.WithLimit.
	api.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))
	coursePhaseApi.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))

	query := db.New(conn)

	skillsService := skills.NewSkillsService(*query, conn)
	teamsService := teams.NewTeamsService(*query, conn)
	surveyService := survey.NewSurveyService(*query, conn)
	allocationService := allocation.NewAllocationService(*query, conn)
	teaseService := tease.NewTeaseService(*query, conn)
	configService := config.NewConfigService(*query, surveyService)
	copyService := copy.NewCopyService(*query, conn)
	privacyService := privacy.NewTeamsPrivacyService(*query, conn)
	coursePhaseDeletionService := coursePhaseDeletion.NewCoursePhaseDeletionService(*query, conn)

	skills.RegisterRoutes(coursePhaseApi, skillsService, promptSDK.AuthenticationMiddleware)
	teams.RegisterRoutes(coursePhaseApi, teamsService, promptSDK.AuthenticationMiddleware)
	survey.RegisterRoutes(coursePhaseApi, surveyService, promptSDK.AuthenticationMiddleware)
	allocation.RegisterRoutes(coursePhaseApi, allocationService, promptSDK.AuthenticationMiddleware)

	tease.RegisterRoutes(api, teaseService, promptSDK.AuthenticationMiddleware) // some tease endpoint are coursePhase independent
	copy.RegisterRoutes(api, copyService)

	config.RegisterRoutes(coursePhaseApi, configService)

	privacy.RegisterRoutes(api, privacyService)
	coursePhaseDeletion.RegisterRoutes(coursePhaseApi, coursePhaseDeletionService)
	return nil
}
