package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/assessments"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel"
	"github.com/prompt-edu/prompt/servers/assessment/bodyLimit"
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
	err := promptSDK.Bootstrap(promptSDK.ServiceOptions{
		ServiceName:    "assessment",
		BasePath:       "/assessment/api",
		DBEnvPrefix:    "ASSESSMENT",
		DefaultDBPort:  "5435",
		SentryDSNEnv:   "SENTRY_DSN_ASSESSMENT",
		DefaultAddress: "localhost:8085",
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
		},
		RegisterRoutes: registerRoutes,
	})
	if err != nil {
		log.Fatalf("Assessment server stopped: %v", err)
	}
}

func registerRoutes(api, coursePhaseApi *gin.RouterGroup, conn *pgxpool.Pool) error {
	// Bootstrap creates coursePhaseApi before calling back, and Gin snapshots a group's chain when
	// a subgroup is created, so the cap goes on both groups. Routes that need more raise it with
	// bodyLimit.WithLimit.
	api.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))
	coursePhaseApi.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))

	query := db.New(conn)

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
	copy.RegisterRoutes(api.Group("", audit.Describe(copy.AuditCopyAction)), copyService)
	privacy.RegisterRoutes(api, privacyService)
	return nil
}
