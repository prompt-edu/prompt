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
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/ls1intum/prompt-sdk"
	"github.com/ls1intum/prompt2/servers/certificate/config"
	db "github.com/ls1intum/prompt2/servers/certificate/db/sqlc"
	"github.com/ls1intum/prompt2/servers/certificate/generator"
	"github.com/ls1intum/prompt2/servers/certificate/participants"
	"github.com/ls1intum/prompt2/servers/certificate/utils"
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

func initSentry() {
	sentryDsn := promptSDK.GetEnv("SENTRY_DSN_CERTIFICATE", "")
	if sentryDsn == "" {
		log.Info("Sentry DSN not configured, skipping initialization")
		return
	}

	transport := sentry.NewHTTPTransport()
	transport.Timeout = 2 * time.Second

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
		Environment:      promptSDK.GetEnv("ENVIRONMENT", "development"),
		Debug:            false,
		Transport:        transport,
		EnableLogs:       true,
		AttachStacktrace: true,
		SendDefaultPII:   true,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	}); err != nil {
		log.Errorf("Sentry initialization failed: %v", err)
		return
	}

	client := sentry.CurrentHub().Client()
	if client == nil {
		log.Error("Sentry client is nil")
		return
	}

	logHook := sentrylogrus.NewLogHookFromClient(
		[]log.Level{log.InfoLevel, log.WarnLevel},
		client,
	)

	eventHook := sentrylogrus.NewEventHookFromClient(
		[]log.Level{log.ErrorLevel, log.FatalLevel, log.PanicLevel},
		client,
	)

	log.AddHook(logHook)
	log.AddHook(eventHook)

	log.RegisterExitHandler(func() {
		eventHook.Flush(5 * time.Second)
		logHook.Flush(5 * time.Second)
	})

	log.Info("Sentry initialized successfully")
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
	initSentry()
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

	api := router.Group("certificate/api/course_phase/:coursePhaseID")
	initKeycloak()

	api.GET("/hello", helloCertificate)

	// Initialize modules
	config.InitConfigModule(api, *query, conn)
	participants.InitParticipantsModule(api, *query)
	generator.InitGeneratorModule(api, *query)

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
