package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/intro_course/config"
	"github.com/prompt-edu/prompt/servers/intro_course/copy"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	"github.com/prompt-edu/prompt/servers/intro_course/developerProfile"
	"github.com/prompt-edu/prompt/servers/intro_course/infrastructureSetup"
	"github.com/prompt-edu/prompt/servers/intro_course/seatPlan"
	"github.com/prompt-edu/prompt/servers/intro_course/tutor"
	log "github.com/sirupsen/logrus"
)

func getDatabaseURL() string {
	dbUser := sdkUtils.GetEnv("DB_USER", "prompt-postgres")
	dbPassword := sdkUtils.GetEnv("DB_PASSWORD", "prompt-postgres")
	dbHost := sdkUtils.GetEnv("DB_HOST_INTRO_COURSE", "localhost")
	dbPort := sdkUtils.GetEnv("DB_PORT_INTRO_COURSE", "5433")
	dbName := sdkUtils.GetEnv("DB_NAME", "prompt")
	sslMode := sdkUtils.GetEnv("SSL_MODE", "disable")
	timeZone := sdkUtils.GetEnv("DB_TIMEZONE", "Europe/Berlin") // Add a timezone parameter

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s", dbUser, dbPassword, dbHost, dbPort, dbName, sslMode, timeZone)
}

func runMigrations(databaseURL string) {
	cmd := exec.Command("migrate", "-path", "./db/migration", "-database", databaseURL, "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}

func initSentry() {
	sentryDsn := sdkUtils.GetEnv("SENTRY_DSN_INTRO_COURSE", "")
	if sentryDsn == "" {
		log.Info("Sentry DSN not configured, skipping initialization")
		return
	}

	transport := sentry.NewHTTPTransport()
	transport.Timeout = 2 * time.Second

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
		Environment:      sdkUtils.GetEnv("ENVIRONMENT", "development"),
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
	baseURL := sdkUtils.GetEnv("KEYCLOAK_HOST", "http://localhost:8081")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "https://" + baseURL
	}

	realm := sdkUtils.GetEnv("KEYCLOAK_REALM_NAME", "prompt")
	coreURL := sdkUtils.GetCoreUrl()
	err := promptSDK.InitAuthenticationMiddleware(baseURL, realm, coreURL)
	if err != nil {
		log.Fatalf("Failed to initialize keycloak: %v", err)
	}
}

// @title           PROMPT Intro Course API
// @version         1.0
// @description     This is the intro course server of PROMPT.
// @host            localhost:8082
// @BasePath        /intro-course/api
// @externalDocs.description  PROMPT Documentation
// @externalDocs.url          https://prompt-edu.github.io/prompt/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token authentication. Use format: Bearer {token}
func main() {
	initSentry()
	defer sentry.Flush(2 * time.Second)

	databaseURL := getDatabaseURL()
	log.Debug("Connecting to database at:", databaseURL)

	runMigrations(databaseURL)

	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	query := db.New(conn)

	router := gin.Default()
	localHost := "http://localhost:3000"
	clientHost := sdkUtils.GetEnv("CORE_HOST", localHost)
	router.Use(sdkUtils.CORS(clientHost))
	router.Use(sentrygin.New(sentrygin.Options{}))

	api := router.Group("intro-course/api/course_phase/:coursePhaseID")
	initKeycloak()
	developerProfile.InitDeveloperProfileModule(api, *query, conn)
	tutor.InitTutorModule(api, *query, conn)
	seatPlan.InitSeatPlanModule(api, *query, conn)

	// Infrastructure Setup
	gitlabAccessToken := sdkUtils.GetEnv("GITLAB_ACCESS_TOKEN", "")
	infrastructureSetup.InitInfrastructureModule(api, *query, conn, gitlabAccessToken)

	copyApi := router.Group("intro-course/api")
	copy.InitCopyModule(copyApi, *query, conn)

	config.InitConfigModule(api, *query, conn)

	serverAddress := sdkUtils.GetEnv("SERVER_ADDRESS", "localhost:8082")
	log.Info("Intro Course Server started")
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
