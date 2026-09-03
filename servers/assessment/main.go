package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/assessments"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel"
	"github.com/prompt-edu/prompt/servers/assessment/categories"
	"github.com/prompt-edu/prompt/servers/assessment/competencies"
	"github.com/prompt-edu/prompt/servers/assessment/copy"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem"
	"github.com/prompt-edu/prompt/servers/assessment/privacy"
	"github.com/prompt-edu/prompt/servers/assessment/schemaModification"
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
	sentryEnabled := promptSDK.GetEnv("SENTRY_ENABLED", "false") == "true"
	if sentryEnabled {
		_ = sdkUtils.InitSentry(promptSDK.GetEnv("SENTRY_DSN_ASSESSMENT", ""))
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

	api := router.Group("/assessment/api")

	// The audit auto-capture middleware must wrap every module's routes. Gin
	// snapshots the middleware chain when a route or subgroup is registered, so
	// it has to be registered before coursePhaseApi and all module routes below;
	// otherwise those routes keep the pre-audit chain and their mutations are
	// never captured. It is a no-op unless the AUDIT_ENABLED toggle is set.
	api.Use(audit.Middleware(audit.NewCoreSink(sdkUtils.GetCoreUrl(), "assessment")))

	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")

	if err := promptSDK.InitPhaseKeycloak(); err != nil {
		log.Fatalf("Failed to initialize keycloak: %v", err)
	}

	coursePhaseApi.GET("/hello", helloAssessment)

	assessmentSchemaService := assessmentSchemas.NewAssessmentSchemaService(*query, conn)
	coursePhaseConfigService := coursePhaseConfig.NewCoursePhaseConfigService(*query, conn, assessmentSchemaService)
	schemaModificationService := schemaModification.NewSchemaModificationService(assessmentSchemaService, coursePhaseConfigService, *query)

	competencyService := competencies.NewCompetencyService(*query, conn, assessmentSchemaService, schemaModificationService)
	categoryService := categories.NewCategoryService(*query, conn, assessmentSchemaService, schemaModificationService, coursePhaseConfigService)

	evaluationCompletionService := evaluationCompletion.NewEvaluationCompletionService(*query, conn, coursePhaseConfig.GetTeamsForCoursePhase, coursePhaseConfigService)
	evaluationService := evaluations.NewEvaluationService(*query, conn, evaluationCompletionService, coursePhaseConfigService)
	feedbackItemService := feedbackItem.NewFeedbackItemService(*query, conn, evaluationCompletionService)

	assessmentCompletionService := assessmentCompletion.NewAssessmentCompletionService(*query, conn, coursePhaseConfigService)
	categoryAssessmentService := categoryAssessment.NewCategoryAssessmentService(*query, conn, assessmentCompletionService)
	actionItemService := actionItem.NewActionItemService(*query, assessmentCompletionService, coursePhaseConfigService)
	scoreLevelService := scoreLevel.NewScoreLevelService(*query)
	assessmentService := assessments.NewAssessmentService(*query, conn, assessmentCompletionService, categoryAssessmentService, actionItemService, scoreLevelService, evaluationService, coursePhaseConfigService)

	competencies.RegisterRoutes(coursePhaseApi, competencyService, promptSDK.AuthenticationMiddleware)
	categories.RegisterRoutes(coursePhaseApi, categoryService, promptSDK.AuthenticationMiddleware)

	coursePhaseConfig.RegisterRoutes(coursePhaseApi, coursePhaseConfigService, promptSDK.AuthenticationMiddleware)
	assessmentSchemas.RegisterRoutes(coursePhaseApi, assessmentSchemaService, promptSDK.AuthenticationMiddleware)

	assessments.RegisterRoutes(coursePhaseApi, assessmentService, coursePhaseConfigService, promptSDK.AuthenticationMiddleware)
	assessmentCompletion.RegisterRoutes(coursePhaseApi, assessmentCompletionService, coursePhaseConfigService, promptSDK.AuthenticationMiddleware)
	categoryAssessment.RegisterRoutes(coursePhaseApi, categoryAssessmentService, coursePhaseConfigService, promptSDK.AuthenticationMiddleware)
	actionItem.RegisterRoutes(coursePhaseApi, actionItemService, coursePhaseConfigService, promptSDK.AuthenticationMiddleware)
	scoreLevel.RegisterRoutes(coursePhaseApi, scoreLevelService, promptSDK.AuthenticationMiddleware)
	evaluations.RegisterRoutes(coursePhaseApi, evaluationService, promptSDK.AuthenticationMiddleware)
	evaluationCompletion.RegisterRoutes(coursePhaseApi, evaluationCompletionService, promptSDK.AuthenticationMiddleware)
	feedbackItem.RegisterRoutes(coursePhaseApi, feedbackItemService, promptSDK.AuthenticationMiddleware)

	copyService := copy.NewCopyService(*query, conn)
	privacyService := privacy.NewPrivacyService(*query, conn)

	// The SDK registrar owns the POST /copy route itself and has no per-route
	// slot for the audit label, so it is attached through a group. An empty
	// relative path leaves the registered route path unchanged.
	copy.RegisterRoutes(api.Group("", audit.Describe("Copied assessment phase")), copyService, promptSDK.AuthenticationMiddleware)
	privacy.RegisterRoutes(api, privacyService)

	promptTypes.RegisterInfoEndpoint(api, promptTypes.ServiceInfo{
		ServiceName: "assessment",
		Version:     promptSDK.GetEnv("SERVER_IMAGE_TAG", ""),
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
			promptTypes.CapabilityAuditLog:        audit.Enabled(),
		},
	}, func() bool {
		return conn.Ping(context.Background()) == nil
	})

	serverAddress := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8085")
	log.Info("Assessment Server started")
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
