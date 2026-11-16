package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/interview/bodyLimit"
	"github.com/prompt-edu/prompt/servers/interview/config"
	"github.com/prompt-edu/prompt/servers/interview/copy"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interview_assignment "github.com/prompt-edu/prompt/servers/interview/interviewAssignment"
	interview_review "github.com/prompt-edu/prompt/servers/interview/interviewReview"
	interview_slot "github.com/prompt-edu/prompt/servers/interview/interviewSlot"
	"github.com/prompt-edu/prompt/servers/interview/privacy"
	log "github.com/sirupsen/logrus"
)

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

func main() {
	err := promptSDK.Bootstrap(promptSDK.ServiceOptions{
		ServiceName:    "interview",
		BasePath:       "/interview/api",
		DBEnvPrefix:    "INTERVIEW",
		DefaultDBPort:  "5438",
		SentryDSNEnv:   "SENTRY_DSN_INTERVIEW",
		DefaultAddress: "localhost:8087",
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
		},
		RegisterRoutes: registerRoutes,
	})
	if err != nil {
		log.Fatalf("Interview server stopped: %v", err)
	}
}

func registerRoutes(api, coursePhaseApi *gin.RouterGroup, conn *pgxpool.Pool) error {
	// Bootstrap creates coursePhaseApi before calling back, and Gin snapshots a group's chain when
	// a subgroup is created, so the cap goes on both groups. Routes that need more raise it with
	// bodyLimit.WithLimit.
	api.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))
	coursePhaseApi.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))

	query := db.New(conn)

	coursePhaseApi.GET("/hello", promptSDK.AuthenticationMiddleware(
		promptSDK.PromptAdmin,
		promptSDK.PromptLecturer,
		promptSDK.CourseLecturer,
		promptSDK.CourseEditor,
		promptSDK.CourseStudent,
	), helloInterviewServer)

	privacyService := privacy.NewPrivacyService(*query, conn)
	interviewSlotService := interview_slot.NewInterviewSlotService(*query, conn)
	interviewAssignmentService := interview_assignment.NewInterviewAssignmentService(*query, conn)
	interviewReviewService := interview_review.NewInterviewReviewService(*query)

	copy.RegisterRoutes(api)
	privacy.RegisterRoutes(api, privacyService)

	config.RegisterRoutes(coursePhaseApi)

	interview_slot.RegisterRoutes(coursePhaseApi, interviewSlotService, promptSDK.AuthenticationMiddleware)
	interview_assignment.RegisterRoutes(coursePhaseApi, interviewAssignmentService, promptSDK.AuthenticationMiddleware)
	interview_review.RegisterRoutes(coursePhaseApi, interviewReviewService, promptSDK.AuthenticationMiddleware)
	return nil
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
