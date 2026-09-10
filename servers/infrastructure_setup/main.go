package main

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/config"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/copy"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/coursePhaseDeletion"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/encryption"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/execution"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/phaseconfig"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/privacy"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/providerconfig"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/resourceconfig"
	log "github.com/sirupsen/logrus"
)

var (
	dbUser     = promptSDK.GetEnv("DB_USER", "prompt-postgres")
	dbPassword = promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
	dbHost     = promptSDK.GetEnv("DB_HOST_INFRASTRUCTURE_SETUP", "localhost")
	dbPort     = promptSDK.GetEnv("DB_PORT_INFRASTRUCTURE_SETUP", "5441")
	dbName     = promptSDK.GetEnv("DB_NAME", "prompt")
	sslMode    = promptSDK.GetEnv("SSL_MODE", "disable")
	timeZone   = promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin")
)

func getDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, sslMode, timeZone)
}

// registerProviderFactories wires provider constructors into the execution registry.
func registerProviderFactories() {
	makeFactory := func(pt string) func([]byte) (provider.Provider, error) {
		return func(creds []byte) (provider.Provider, error) {
			return providerconfig.BuildProviderFromEncryptedCreds(pt, creds)
		}
	}

	execution.Registry["gitlab"] = makeFactory("gitlab")
	execution.Registry["slack"] = makeFactory("slack")
	execution.Registry["outline"] = makeFactory("outline")
	execution.Registry["rancher"] = makeFactory("rancher")
	execution.Registry["keycloak"] = makeFactory("keycloak")
}

// @title           PROMPT Infrastructure Setup API
// @version         1.0
// @description     Manages infrastructure resource provisioning for course phases.
// @host            localhost:8091
// @BasePath        /infrastructure-setup/api
// @externalDocs.description  PROMPT Documentation
// @externalDocs.url          https://prompt-edu.github.io/prompt/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token authentication. Use format: Bearer {token}
func main() {
	_ = sdkUtils.InitSentry(promptSDK.GetEnv("SENTRY_DSN_INFRASTRUCTURE_SETUP", ""))
	defer sentry.Flush(2 * time.Second)

	databaseURL := getDatabaseURL()
	log.Debugf("Connecting to database at host=%s port=%s db=%s user=%s sslmode=%s",
		dbHost, dbPort, dbName, dbUser, sslMode)

	if err := sdkUtils.RunMigrations(databaseURL, "./db/migration"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer conn.Close()

	if err := encryption.ValidateKey(); err != nil {
		log.Fatalf("Invalid ENCRYPTION_KEY: %v", err)
	}

	if err := promptSDK.InitPhaseKeycloak(); err != nil {
		log.Fatalf("Failed to initialize keycloak: %v", err)
	}
	registerProviderFactories()

	clientHost := promptSDK.GetEnv("CORE_HOST", "http://localhost:3000")

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{}))
	router.Use(promptSDK.CORSMiddleware(clientHost))

	authMw := promptSDK.AuthenticationMiddleware

	// Phase-scoped API routes. These carry external provider credentials, so they are
	// restricted to admins and lecturers - editors are deliberately excluded.
	api := router.Group("infrastructure-setup/api/course_phase/:coursePhaseID",
		authMw(promptSDK.PromptAdmin, promptSDK.CourseLecturer))

	providerconfig.RegisterRoutes(api, providerconfig.NewService(conn))
	resourceconfig.RegisterRoutes(api, resourceconfig.NewService(conn))
	phaseconfig.RegisterRoutes(api, phaseconfig.NewService(conn))
	executionService := execution.NewService(conn)
	execution.RegisterRoutes(api, executionService)

	// Recover instances a crashed process left claimed, now and periodically.
	executionService.StartStaleClaimSweeper(context.Background())

	// Phase deletion. The SDK protects the endpoint itself, and it reads the phase from
	// the path, so it goes on a phase-scoped group without a group-level middleware.
	deletionApi := router.Group("infrastructure-setup/api/course_phase/:coursePhaseID")
	coursePhaseDeletion.RegisterRoutes(deletionApi, coursePhaseDeletion.NewCoursePhaseDeletionService(conn))

	// Copy endpoint (global, not phase-scoped). The config endpoint is phase-scoped but
	// brings its own auth middleware from the SDK, so it gets a group without one.
	copyApi := router.Group("infrastructure-setup/api")
	configApi := router.Group("infrastructure-setup/api/course_phase/:coursePhaseID")
	copy.RegisterRoutes(copyApi, copy.NewService(conn), authMw)
	config.RegisterRoutes(configApi, config.NewService(conn), authMw)

	// Privacy export and deletion. Core addresses these per phase type at
	// <base URL>/privacy/..., and both bring their own middleware from the SDK.
	privacy.RegisterRoutes(copyApi, privacy.NewPrivacyService(conn))

	// Public /info endpoint consumed by the management console's System Status page.
	promptTypes.RegisterInfoEndpoint(copyApi, promptTypes.ServiceInfo{
		ServiceName: "infrastructure-setup",
		Version:     promptSDK.GetEnv("SERVER_IMAGE_TAG", ""),
		Capabilities: map[string]bool{
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
			promptTypes.CapabilityPhaseDeletion:   true,
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
		},
	}, func() bool {
		// A health probe must answer even when the database does not: bound the ping
		// rather than letting /info hang on an unreachable database.
		pingCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		return conn.Ping(pingCtx) == nil
	})

	// Health check.
	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	serverAddress := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8091")
	log.Info("Infrastructure Setup Server started on ", serverAddress)
	if err := router.Run(serverAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
