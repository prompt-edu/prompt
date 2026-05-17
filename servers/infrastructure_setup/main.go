package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/copy"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/execution"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/phaseconfig"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/providerconfig"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/resourceconfig"
	log "github.com/sirupsen/logrus"
)

var (
	dbUser     = promptSDK.GetEnv("DB_USER", "prompt-postgres")
	dbPassword = promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
	dbHost     = promptSDK.GetEnv("DB_HOST_INFRASTRUCTURE_SETUP", "localhost")
	dbPort     = promptSDK.GetEnv("DB_PORT_INFRASTRUCTURE_SETUP", "5440")
	dbName     = promptSDK.GetEnv("DB_NAME", "prompt")
	sslMode    = promptSDK.GetEnv("SSL_MODE", "disable")
	timeZone   = promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin")
)

func getDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, sslMode, timeZone)
}

func sanitizeDatabaseURL(input string) string {
	return strings.ReplaceAll(input, dbPassword, "***")
}

func runMigrations(databaseURL string) {
	cmd := exec.Command("migrate", "-path", "./db/migration", "-database", databaseURL, "up")
	output, err := cmd.CombinedOutput()
	sanitized := sanitizeDatabaseURL(string(output))
	if err != nil {
		log.Fatalf("Failed to run migrations: %v\n%s", err, sanitized)
	}
	fmt.Print(sanitized)
}

func initKeycloak() {
	baseURL := promptSDK.GetEnv("KEYCLOAK_HOST", "http://localhost:8081")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "https://" + baseURL
	}
	realm := promptSDK.GetEnv("KEYCLOAK_REALM_NAME", "prompt")
	coreURL := sdkUtils.GetCoreUrl()
	if err := promptSDK.InitAuthenticationMiddleware(baseURL, realm, coreURL); err != nil {
		log.Fatalf("Failed to initialize keycloak: %v", err)
	}
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
// @host            localhost:8089
// @BasePath        /infrastructure-setup/api
func main() {
	_ = sdkUtils.InitSentry(promptSDK.GetEnv("SENTRY_DSN_INFRASTRUCTURE_SETUP", ""))
	defer sentry.Flush(2 * time.Second)

	databaseURL := getDatabaseURL()
	log.Debugf("Connecting to database at host=%s port=%s db=%s user=%s sslmode=%s",
		dbHost, dbPort, dbName, dbUser, sslMode)

	runMigrations(databaseURL)

	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer conn.Close()

	// Recover instances stuck in_progress from a previous crash.
	queries := db.New(conn)
	if err := queries.ResetInProgressToPending(context.Background()); err != nil {
		log.WithError(err).Warn("startup recovery: ResetInProgressToPending failed")
	}

	initKeycloak()
	registerProviderFactories()

	clientHost := promptSDK.GetEnv("CORE_HOST", "http://localhost:3000")

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{}))
	router.Use(promptSDK.CORSMiddleware(clientHost))

	authMw := promptSDK.AuthenticationMiddleware

	// Phase-scoped API routes (lecturer + editor access).
	api := router.Group("infrastructure-setup/api/course_phase/:coursePhaseID",
		authMw(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor))

	providerconfig.RegisterRoutes(api, providerconfig.NewService(conn))
	resourceconfig.RegisterRoutes(api, resourceconfig.NewService(conn))
	phaseconfig.RegisterRoutes(api, phaseconfig.NewService(conn))
	execution.RegisterRoutes(api, execution.NewService(conn))

	// Copy endpoint (global, not phase-scoped).
	copyApi := router.Group("infrastructure-setup/api")
	copy.InitCopyModule(copyApi, conn)

	// Health check.
	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	serverAddress := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8089")
	log.Info("Infrastructure Setup Server started on ", serverAddress)
	if err := router.Run(serverAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
