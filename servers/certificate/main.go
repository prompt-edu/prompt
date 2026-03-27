package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/certificate/config"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/prompt-edu/prompt/servers/certificate/generator"
	"github.com/prompt-edu/prompt/servers/certificate/participants"
	"github.com/prompt-edu/prompt/servers/certificate/privacy"
	"github.com/prompt-edu/prompt/servers/certificate/utils"
	log "github.com/sirupsen/logrus"
)

func getDatabaseURL() string {
	dbUser := promptSDK.GetEnv("DB_USER", "prompt-postgres")
	dbPassword := promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
	dbHost := promptSDK.GetEnv("DB_HOST_CERTIFICATE", "localhost")
	dbPort := promptSDK.GetEnv("DB_PORT_CERTIFICATE", "5439")
	dbName := promptSDK.GetEnv("DB_NAME", "prompt")
	sslMode := promptSDK.GetEnv("SSL_MODE", "disable")
	timeZone := promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s", dbUser, dbPassword, dbHost, dbPort, dbName, sslMode, timeZone)
}

func sanitizeDatabaseURL(input string) string {
	dbPassword := promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
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
		log.Warn("Keycloak host does not start with http(s). Adding https:// as prefix.")
		baseURL = "https://" + baseURL
	}

	realm := promptSDK.GetEnv("KEYCLOAK_REALM_NAME", "prompt")
	coreURL := utils.GetCoreUrl()

	err := promptSDK.InitAuthenticationMiddleware(baseURL, realm, coreURL)
	if err != nil {
		log.Fatalf("Failed to initialize keycloak: %v", err)
	}
}

func main() {
	_ = sdkUtils.InitSentry(promptSDK.GetEnv("SENTRY_DSN_CERTIFICATE", ""))
	defer sentry.Flush(2 * time.Second)

	databaseURL := getDatabaseURL()
	log.Debug("Connecting to database")

	runMigrations(databaseURL)

	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	query := db.New(conn)

	clientHost := promptSDK.GetEnv("CORE_HOST", "http://localhost:3000")

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{}))
	router.Use(promptSDK.CORSMiddleware(clientHost))

	api := router.Group("certificate/api")
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")
	initKeycloak()

	coursePhaseApi.GET("/hello", helloCertificate)

	// Initialize modules
	config.InitConfigModule(coursePhaseApi, *query, conn)
	participants.InitParticipantsModule(coursePhaseApi, *query)
	generator.InitGeneratorModule(coursePhaseApi, *query)
	privacy.InitPrivacyModule(api, *query, conn)

	serverAddress := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8088")
	log.Info("Certificate Server started")
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func helloCertificate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello from certificate service",
	})
}
