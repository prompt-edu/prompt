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
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/assessments"
	"github.com/prompt-edu/prompt/servers/assessment/categories"
	"github.com/prompt-edu/prompt/servers/assessment/competencies"
	"github.com/prompt-edu/prompt/servers/assessment/copy"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations"
	"github.com/prompt-edu/prompt/servers/assessment/privacy"
	log "github.com/sirupsen/logrus"
)

func getDatabaseURL() string {
	dbUser := promptSDK.GetEnv("DB_USER", "prompt-postgres")
	dbPassword := promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres")
	dbHost := promptSDK.GetEnv("DB_HOST_ASSESSMENT", "localhost")
	dbPort := promptSDK.GetEnv("DB_PORT_ASSESSMENT", "5435")
	dbName := promptSDK.GetEnv("DB_NAME", "prompt")
	sslMode := promptSDK.GetEnv("SSL_MODE", "disable")
	timeZone := promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin") // Add a timezone parameter

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

// helloAssessment godoc
// @Summary Assessment service health check
// @Description Returns a simple hello message from the assessment service.
// @Tags health
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/hello [get]
func helloAssessment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello from assessment service",
	})
}

// @title           PROMPT Assessment API
// @version         1.0
// @description     This is the assessment server of PROMPT.

// @host      localhost:8085
// @BasePath  /assessment/api

// @externalDocs.description  PROMPT Documentation
// @externalDocs.url          https://prompt-edu.github.io/prompt/
func main() {
	_ = sdkUtils.InitSentry(promptSDK.GetEnv("SENTRY_DSN_ASSESSMENT", ""))
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

	clientHost := promptSDK.GetEnv("CORE_HOST", "http://localhost:3000")

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{}))
	router.Use(promptSDK.CORSMiddleware(clientHost))

	api := router.Group("assessment/api")
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")
	initKeycloak(*query)

	coursePhaseApi.GET("/hello", helloAssessment)

	competencies.InitCompetencyModule(coursePhaseApi, *query, conn)
	categories.InitCategoryModule(coursePhaseApi, *query, conn)
	coursePhaseConfig.InitCoursePhaseConfigModule(coursePhaseApi, *query, conn)
	assessmentSchemas.InitAssessmentSchemaModule(coursePhaseApi, *query, conn)
	assessments.InitAssessmentModule(coursePhaseApi, *query, conn)
	evaluations.InitEvaluationModule(coursePhaseApi, *query, conn)

	copy.InitCopyModule(api, *query, conn)
	privacy.InitPrivacyModule(api, *query, conn)

	serverAddress := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8085")
	log.Info("Assessment Server started")
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
