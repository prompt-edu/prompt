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
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration"
	"github.com/prompt-edu/prompt/servers/core/course"
	"github.com/prompt-edu/prompt/servers/core/course/copy"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	"github.com/prompt-edu/prompt/servers/core/auth"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseType"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/instructorNote"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/mailing"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/privacy"
	"github.com/prompt-edu/prompt/servers/core/storage/files"
	"github.com/prompt-edu/prompt/servers/core/storage/privacyexport"
	"github.com/prompt-edu/prompt/servers/core/student"
	log "github.com/sirupsen/logrus"
)

func getDatabaseURL() string {
	dbUser := sdkUtils.GetEnv("DB_USER", "prompt-postgres")
	dbPassword := sdkUtils.GetEnv("DB_PASSWORD", "prompt-postgres")
	dbHost := sdkUtils.GetEnv("DB_HOST", "localhost")
	dbPort := sdkUtils.GetEnv("DB_PORT", "5432")
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

func initKeycloak(router *gin.RouterGroup, queries db.Queries) {
	baseURL := sdkUtils.GetEnv("KEYCLOAK_HOST", "http://localhost:8081")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "https://" + baseURL
	}

	realm := sdkUtils.GetEnv("KEYCLOAK_REALM_NAME", "prompt")
	clientID := sdkUtils.GetEnv("KEYCLOAK_CLIENT_ID", "prompt-server")
	clientSecret := sdkUtils.GetEnv("KEYCLOAK_CLIENT_SECRET", "")
	idOfClient := sdkUtils.GetEnv("KEYCLOAK_ID_OF_CLIENT", "a584ca61-fa83-4e95-98b6-c5f3157ae4b4")
	expectedAuthorizedParty := sdkUtils.GetEnv("KEYCLOAK_AUTHORIZED_PARTY", "prompt-client")

	log.Info("Debugging: baseURL: ", baseURL, " realm: ", realm, " clientID: ", clientID, " idOfClient: ", idOfClient, " expectedAuthorizedParty: ", expectedAuthorizedParty)

	// first we initialize the keycloak token verfier
	keycloakTokenVerifier.InitKeycloakTokenVerifier(context.Background(), baseURL, realm, clientID, expectedAuthorizedParty, queries)

	err := keycloakRealmManager.InitKeycloak(context.Background(), router, baseURL, realm, clientID, clientSecret, idOfClient, expectedAuthorizedParty, queries)
	if err != nil {
		log.Error("Failed to initialize keycloak: ", err)
	}
}

func initMailing(router *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	log.Debug("Reading mailing environment variables...")

	clientURL := sdkUtils.GetEnv("CORE_HOST", "localhost:3000") // required for application link in mails
	smtpHost := sdkUtils.GetEnv("SMTP_HOST", "127.0.0.1")
	smtpPort := sdkUtils.GetEnv("SMTP_PORT", "25")
	smtpUsername := sdkUtils.GetEnv("SMTP_USERNAME", "")
	smtpPassword := sdkUtils.GetEnv("SMTP_PASSWORD", "")
	senderEmail := sdkUtils.GetEnv("SENDER_EMAIL", "")
	senderName := sdkUtils.GetEnv("SENDER_NAME", "Prompt Mailing Service")

	log.Debug("Environment variables read:")
	log.Debug("CORE_HOST: ", clientURL)
	log.Debug("SMTP_HOST: ", smtpHost)
	log.Debug("SMTP_PORT: ", smtpPort)
	log.Debug("SMTP_USERNAME: ", smtpUsername)
	log.Debug("SMTP_PASSWORD: ", "[REDACTED]") // Don't log the actual password
	log.Debug("SENDER_EMAIL: ", senderEmail)
	log.Debug("SENDER_NAME: ", senderName)

	log.Info("Initializing mailing service with SMTP host: ", smtpHost, " port: ", smtpPort, " sender email: ", senderEmail)

	mailing.InitMailingModule(router, queries, conn, smtpHost, smtpPort, smtpUsername, smtpPassword, senderName, senderEmail, clientURL)
}

// @title           PROMPT Core API
// @version         1.0
// @description     This is a core sever of PROMPT.

// @host      localhost:8080
// @BasePath  /api/

// @externalDocs.description  PROMPT Documentation
// @externalDocs.url          https://prompt-edu.github.io/prompt/
func main() {
	if sdkUtils.GetEnv("DEBUG", "false") == "true" {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug mode is enabled")
	}

	// initialize Sentry
	sentryEnabled := sdkUtils.GetEnv("SENTRY_ENABLED", "false") == "true"
	if sentryEnabled {
		_ = sdkUtils.InitSentry(sdkUtils.GetEnv("SENTRY_DSN_CORE", ""))
		defer sentry.Flush(2 * time.Second) // Flush buffered events before exiting (2 seconds timeout)
	}

	// establish database connection
	databaseURL := getDatabaseURL()
	log.Debug("Connecting to database at:", databaseURL)

	// run migrations
	runMigrations(databaseURL)

	// establish db connection
	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	query := db.New(conn)

	router := gin.Default()
	if sentryEnabled {
		router.Use(sentrygin.New(sentrygin.Options{}))
	}
	localHost := "http://localhost:3000"
	clientHost := sdkUtils.GetEnv("CORE_HOST", localHost)
	router.Use(sdkUtils.CORS(clientHost))

	api := router.Group("/api")
	api.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	initKeycloak(api, *query)
	permissionValidation.InitValidationService(*query, conn)

	// this initializes also all available course phase types
	environment := sdkUtils.GetEnv("ENVIRONMENT", "development")
	isDevEnvironment := environment == "development"
	coursePhaseType.InitCoursePhaseTypeModule(api, *query, conn, isDevEnvironment)

	coreHost := sdkUtils.GetEnv("CORE_HOST", "localhost:8080")
	resolution.InitResolutionModule(coreHost)

	auth.InitAuthModule(api, *query, conn)
	initMailing(api, *query, conn)
	student.InitStudentModule(api, *query, conn)
	course.InitCourseModule(api, *query, conn)
	copy.InitCourseCopyModule(api, *query, conn)
	coursePhase.InitCoursePhaseModule(api, *query, conn)
	courseParticipation.InitCourseParticipationModule(api, *query, conn)
	coursePhaseParticipation.InitCoursePhaseParticipationModule(api, *query, conn)
	applicationAdministration.InitApplicationAdministrationModule(api, *query, conn)
	instructorNote.InitInstructorNoteModule(api, *query, conn)

	if err := files.Init(*query, conn); err != nil {
		log.Fatalf("Failed to initialize prompt file storage: %v", err)
	}

	if err := privacyexport.Init(); err != nil {
		log.Fatalf("Failed to initialize privacy export storage: %v", err)
	}

	privacy.InitPrivacyModule(api, *query, conn)

	serverAddress := sdkUtils.GetEnv("SERVER_ADDRESS", "localhost:8080")
	log.Info("Core Server started")
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
