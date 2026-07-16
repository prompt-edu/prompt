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
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/example_server/config"
	"github.com/prompt-edu/prompt/servers/example_server/copy"
	db "github.com/prompt-edu/prompt/servers/example_server/db/sqlc"
	"github.com/prompt-edu/prompt/servers/example_server/example"
	log "github.com/sirupsen/logrus"
)

var dbUser string = promptSDK.GetEnv("DB_USER", "prompt-postgres")
var dbPassword string = promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
var dbHost string = promptSDK.GetEnv("DB_HOST_EXAMPLE_SERVER", "localhost")
var dbPort string = promptSDK.GetEnv("DB_PORT_EXAMPLE_SERVER", "5437")
var dbName string = promptSDK.GetEnv("DB_NAME", "prompt")
var sslMode string = promptSDK.GetEnv("SSL_MODE", "disable")
var timeZone string = promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin") // Add a timezone parameter

func getDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s", dbUser, dbPassword, dbHost, dbPort, dbName, sslMode, timeZone)
}

func sanitizeDatabaseURL(input string) string {
	dbPassword := promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
	if dbPassword == "" {
		return input
	}
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
func initKeycloak(queries db.Queries) {
	baseURL := promptSDK.GetEnv("KEYCLOAK_HOST", "http://localhost:8081")
	if !strings.HasPrefix(baseURL, "http") {
		log.Warn("Keycloak host does not start with http(s). Adding https:// as prefix.")
		baseURL = "https://" + baseURL
	}

	realm := promptSDK.GetEnv("KEYCLOAK_REALM_NAME", "prompt")

	coreURL := sdkUtils.GetCoreUrl()
	err := promptSDK.InitAuthenticationMiddleware(baseURL, realm, coreURL)
	if err != nil {
		log.Fatalf("Failed to initialize keycloak: %v", err)
	}
}

// @title           PROMPT Example API
// @version         1.0
// @description     This is the example server of PROMPT.
// @host            localhost:8086
// @BasePath        /example-service/api
// @externalDocs.description  PROMPT Documentation
// @externalDocs.url          https://prompt-edu.github.io/prompt/
func main() {
	sentryEnabled := promptSDK.GetEnv("SENTRY_ENABLED", "false") == "true"
	if sentryEnabled {
		_ = sdkUtils.InitSentry(promptSDK.GetEnv("SENTRY_DSN_EXAMPLE_SERVER", ""))
		defer sentry.Flush(2 * time.Second)
	}

	databaseURL := getDatabaseURL()
	log.Debugf("Connecting to database at host=%s port=%s db=%s user=%s sslmode=%s", dbHost, dbPort, dbName, dbUser, sslMode)

	runMigrations(databaseURL)

	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer conn.Close()

	query := db.New(conn)

	clientHost := promptSDK.GetEnv("CORE_HOST", "http://localhost:3000")

	router := gin.Default()
	if sentryEnabled {
		router.Use(sentrygin.New(sentrygin.Options{}))
	}
	router.Use(promptSDK.CORSMiddleware(clientHost))

	api := router.Group("example-service/api/course_phase/:coursePhaseID")
	initKeycloak(*query)

	api.GET("/hello", helloExampleServer)

	copyApi := router.Group("example-service/api")
	copy.InitCopyModule(copyApi, *query, conn)

	promptTypes.RegisterInfoEndpoint(copyApi, promptTypes.ServiceInfo{
		ServiceName: "example-service",
		Version:     promptSDK.GetEnv("SERVER_IMAGE_TAG", ""),
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   false,
			promptTypes.CapabilityPrivacyDeletion: false,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
		},
	}, func() bool {
		ctt, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		return conn.Ping(ctt) == nil
	})

	config.InitConfigModule(api, *query, conn)

	example.InitExampleModule(api, *query, conn)

	serverAddress := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8086")
	log.Info("Example Server started")
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// helloExampleServer godoc
// @Summary Example server health check
// @Description Returns a simple hello message from the example server.
// @Tags example
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/hello [get]
func helloExampleServer(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello from the example service",
	})
}
