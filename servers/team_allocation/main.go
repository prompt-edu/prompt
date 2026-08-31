package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation"
	"github.com/prompt-edu/prompt/servers/team_allocation/config"
	"github.com/prompt-edu/prompt/servers/team_allocation/copy"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/privacy"
	"github.com/prompt-edu/prompt/servers/team_allocation/skills"
	"github.com/prompt-edu/prompt/servers/team_allocation/survey"
	teams "github.com/prompt-edu/prompt/servers/team_allocation/team"
	"github.com/prompt-edu/prompt/servers/team_allocation/tease"
	log "github.com/sirupsen/logrus"
)

func getDatabaseURL() string {
	dbUser := promptSDK.GetEnv("DB_USER", "prompt-postgres")
	dbPassword := promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
	dbHost := promptSDK.GetEnv("DB_HOST_TEAM_ALLOCATION", "localhost")
	dbPort := promptSDK.GetEnv("DB_PORT_TEAM_ALLOCATION", "5434")
	dbName := promptSDK.GetEnv("DB_NAME", "prompt")
	sslMode := promptSDK.GetEnv("SSL_MODE", "disable")
	timeZone := promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin") // Add a timezone parameter

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s", dbUser, dbPassword, dbHost, dbPort, dbName, sslMode, timeZone)
}

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
	sentryEnabled := promptSDK.GetEnv("SENTRY_ENABLED", "false") == "true"
	if sentryEnabled {
		_ = sdkUtils.InitSentry(promptSDK.GetEnv("SENTRY_DSN_TEAM_ALLOCATION", ""))
		defer sentry.Flush(2 * time.Second)
	}

	databaseURL := getDatabaseURL()
	log.Debug("Connecting to database at:", databaseURL)

	if err := sdkUtils.RunMigrations(databaseURL, "./db/migration"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	query := db.New(conn)

	clientHost := promptSDK.GetEnv("CORE_HOST", "http://localhost:3000")

	router := gin.Default()
	if sentryEnabled {
		router.Use(sentrygin.New(sentrygin.Options{}))
	}
	router.Use(promptSDK.CORSMiddleware(clientHost))

	api := router.Group("/team-allocation/api")
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")
	if err := promptSDK.InitPhaseKeycloak(); err != nil {
		log.Fatalf("Failed to initialize keycloak: %v", err)
	}

	// No health endpoint; health checks are handled externally.

	skillsService := skills.NewSkillsService(*query, conn)
	teamsService := teams.NewTeamsService(*query, conn)
	surveyService := survey.NewSurveyService(*query, conn)
	allocationService := allocation.NewAllocationService(*query)
	teaseService := tease.NewTeaseService(*query, conn)
	configService := config.NewConfigService(*query, surveyService)
	copyService := copy.NewCopyService(*query, conn)
	privacyService := privacy.NewTeamsPrivacyService(*query, conn)

	skills.RegisterRoutes(coursePhaseApi, skillsService, promptSDK.AuthenticationMiddleware)
	teams.RegisterRoutes(coursePhaseApi, teamsService, promptSDK.AuthenticationMiddleware)
	survey.RegisterRoutes(coursePhaseApi, surveyService, promptSDK.AuthenticationMiddleware)
	allocation.RegisterRoutes(coursePhaseApi, allocationService, promptSDK.AuthenticationMiddleware)

	tease.RegisterRoutes(router.Group("team-allocation/api"), teaseService, promptSDK.AuthenticationMiddleware) // some tease endpoint are coursePhase independent

	copyApi := router.Group("team-allocation/api")
	copy.RegisterRoutes(copyApi, copyService, promptSDK.AuthenticationMiddleware)

	config.RegisterRoutes(coursePhaseApi, configService, promptSDK.AuthenticationMiddleware)

	privacy.RegisterRoutes(api, privacyService)

	promptTypes.RegisterInfoEndpoint(copyApi, promptTypes.ServiceInfo{
		ServiceName: "team-allocation",
		Version:     promptSDK.GetEnv("SERVER_IMAGE_TAG", ""),
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
		},
	}, func() bool {
		return conn.Ping(context.Background()) == nil
	})

	serverAddress := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8083")
	log.Info("Team Allocation Server started")
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
