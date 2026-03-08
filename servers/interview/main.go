package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/interview/config"
	"github.com/prompt-edu/prompt/servers/interview/copy"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interview_assignment "github.com/prompt-edu/prompt/servers/interview/interviewAssignment"
	interview_slot "github.com/prompt-edu/prompt/servers/interview/interviewSlot"
	log "github.com/sirupsen/logrus"
)

var dbUser string = promptSDK.GetEnv("DB_USER", "prompt-postgres")
var dbPassword string = promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
var dbHost string = promptSDK.GetEnv("DB_HOST_INTERVIEW", "localhost")
var dbPort string = promptSDK.GetEnv("DB_PORT_INTERVIEW", "5438")
var dbName string = promptSDK.GetEnv("DB_NAME", "prompt")
var sslMode string = promptSDK.GetEnv("SSL_MODE", "disable")
var timeZone string = promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin")

func getDatabaseURL() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(dbUser, dbPassword),
		Host:   fmt.Sprintf("%s:%s", dbHost, dbPort),
		Path:   "/" + dbName,
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	q.Set("TimeZone", timeZone)
	u.RawQuery = q.Encode()
	return u.String()
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
	sentryDsn := promptSDK.GetEnv("SENTRY_DSN_INTERVIEW", "")
	if sentryDsn == "" {
		log.Info("Sentry DSN not configured, skipping initialization")
		return
	}

	transport := sentry.NewHTTPTransport()
	transport.Timeout = 2 * time.Second
	sendDefaultPII, err := strconv.ParseBool(promptSDK.GetEnv("SENTRY_SEND_DEFAULT_PII", "false"))
	if err != nil {
		log.Warnf("Invalid SENTRY_SEND_DEFAULT_PII value, defaulting to false: %v", err)
		sendDefaultPII = false
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
		Environment:      promptSDK.GetEnv("ENVIRONMENT", "development"),
		Debug:            false,
		Transport:        transport,
		EnableLogs:       true,
		AttachStacktrace: true,
		SendDefaultPII:   sendDefaultPII,
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

// @title           PROMPT Interview API
// @version         1.0
// @description     This is the interview server of PROMPT.
// @host            localhost:8087
// @BasePath        /interview/api
// @externalDocs.description  PROMPT Documentation
// @externalDocs.url          https://prompt-edu.github.io/prompt/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token authentication. Use format: Bearer {token}
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

func main() {
	initSentry()
	defer sentry.Flush(2 * time.Second)

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
	router.Use(sentrygin.New(sentrygin.Options{}))
	router.Use(promptSDK.CORSMiddleware(clientHost))

	api := router.Group("interview/api/course_phase/:coursePhaseID")
	initKeycloak(*query)

	api.GET("/hello", promptSDK.AuthenticationMiddleware(
		promptSDK.PromptAdmin,
		promptSDK.PromptLecturer,
		promptSDK.CourseLecturer,
		promptSDK.CourseEditor,
		promptSDK.CourseStudent,
	), helloInterviewServer)

	copyApi := router.Group("interview/api")
	copy.InitCopyModule(copyApi, *query, conn)

	promptTypes.RegisterInfoEndpoint(copyApi, promptTypes.ServiceInfo{
		ServiceName: "interview",
		Version:     promptSDK.GetEnv("SERVER_IMAGE_TAG", ""),
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyStudentExport:   false,
			promptTypes.CapabilityPrivacyStudentDeletion: false,
			promptTypes.CapabilityPhaseCopy:              true,
			promptTypes.CapabilityPhaseConfig:            true,
		},
	}, func() bool {
		return conn.Ping(context.Background()) == nil
	})

	config.InitConfigModule(api, *query, conn)

	interview_slot.InitInterviewSlotModule(api, *query, conn)
	interview_assignment.InitInterviewAssignmentModule(api, *query, conn)

	serverAddress := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8087")
	log.Info("Interview Server started")
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// helloInterviewServer godoc
// @Summary Interview service hello endpoint
// @Description Returns a simple response from the interview service
// @Tags interview
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/hello [get]
func helloInterviewServer(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello from the interview service",
	})
}
