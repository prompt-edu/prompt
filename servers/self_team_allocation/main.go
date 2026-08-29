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
	"github.com/prompt-edu/prompt/servers/self_team_allocation/allocation"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/config"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/copy"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/privacy"
	teams "github.com/prompt-edu/prompt/servers/self_team_allocation/team"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/timeframe"

	log "github.com/sirupsen/logrus"
)

func getDatabaseURL() string {
	dbUser := promptSDK.GetEnv("DB_USER", "prompt-postgres")
	dbPassword := promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
	dbHost := promptSDK.GetEnv("DB_HOST_SELF_TEAM_ALLOCATION", "localhost")
	dbPort := promptSDK.GetEnv("DB_PORT_SELF_TEAM_ALLOCATION", "5436")
	dbName := promptSDK.GetEnv("DB_NAME", "prompt")
	sslMode := promptSDK.GetEnv("SSL_MODE", "disable")
	timeZone := promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin") // Add a timezone parameter

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s", dbUser, dbPassword, dbHost, dbPort, dbName, sslMode, timeZone)
}

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
	sentryEnabled := promptSDK.GetEnv("SENTRY_ENABLED", "false") == "true"
	if sentryEnabled {
		_ = sdkUtils.InitSentry(promptSDK.GetEnv("SENTRY_DSN_SELF_TEAM_ALLOCATION", ""))
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

	api := router.Group("self-team-allocation/api")
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")
	if err := promptSDK.InitPhaseKeycloak(); err != nil {
		log.Fatalf("Failed to initialize keycloak: %v", err)
	}

	coursePhaseApi.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello from team self assignment service"})
	})

	timeframeService := timeframe.NewTimeframeService(*query)
	teamsService := teams.NewTeamsService(*query, conn, timeframeService)
	assignmentService := teams.NewAssignmentService(*query)
	allocationService := allocation.NewAllocationService(*query)
	configService := config.NewConfigService(*query)
	privacyService := privacy.NewPrivacyService(*query, conn)

	teams.RegisterRoutes(coursePhaseApi, teamsService, assignmentService, promptSDK.AuthenticationMiddleware)
	timeframe.RegisterRoutes(coursePhaseApi, timeframeService, promptSDK.AuthenticationMiddleware)
	allocation.RegisterRoutes(coursePhaseApi, allocationService, promptSDK.AuthenticationMiddleware)
	copy.RegisterRoutes(api, promptSDK.AuthenticationMiddleware)
	privacy.RegisterRoutes(api, privacyService)

	config.RegisterRoutes(coursePhaseApi, configService, promptSDK.AuthenticationMiddleware)

	promptTypes.RegisterInfoEndpoint(api, promptTypes.ServiceInfo{
		ServiceName: "self-team-allocation",
		Version:     promptSDK.GetEnv("SERVER_IMAGE_TAG", ""),
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
		},
	}, func() bool {
		ctt, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		return conn.Ping(ctt) == nil
	})

	serverAddress := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8084")
	log.Info("Self Team Allocation Server started")
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
