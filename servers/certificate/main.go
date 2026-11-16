package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/certificate/bodyLimit"
	"github.com/prompt-edu/prompt/servers/certificate/config"
	"github.com/prompt-edu/prompt/servers/certificate/copy"
	"github.com/prompt-edu/prompt/servers/certificate/coursePhaseDeletion"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/prompt-edu/prompt/servers/certificate/generator"
	"github.com/prompt-edu/prompt/servers/certificate/participants"
	"github.com/prompt-edu/prompt/servers/certificate/privacy"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := promptSDK.Bootstrap(promptSDK.ServiceOptions{
		ServiceName:    "certificate",
		BasePath:       "/certificate/api",
		DBEnvPrefix:    "CERTIFICATE",
		DefaultDBPort:  "5439",
		SentryDSNEnv:   "SENTRY_DSN_CERTIFICATE",
		DefaultAddress: "localhost:8088",
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
			promptTypes.CapabilityPhaseDeletion:   true,
		},
		RegisterRoutes: registerRoutes,
	})
	if err != nil {
		log.Fatalf("Certificate server stopped: %v", err)
	}
}

func registerRoutes(api, coursePhaseApi *gin.RouterGroup, conn *pgxpool.Pool) error {
	// Bootstrap creates coursePhaseApi before calling back, and Gin snapshots a group's chain when
	// a subgroup is created, so the cap goes on both groups. Routes that need more raise it with
	// bodyLimit.WithLimit.
	api.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))
	coursePhaseApi.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))

	query := db.New(conn)

	coursePhaseApi.GET("/hello", helloCertificate)

	configService := config.NewConfigService(*query)
	participantsService := participants.NewParticipantsService(*query, sdkUtils.GetCoreUrl())
	generatorService := generator.NewGeneratorService(*query, configService, participantsService, participantsService)
	privacyService := privacy.NewPrivacyService(*query)
	copyService := copy.NewCopyService(*query)
	coursePhaseDeletionService := coursePhaseDeletion.NewCoursePhaseDeletionService(*query, conn)

	config.RegisterRoutes(coursePhaseApi, configService, promptSDK.AuthenticationMiddleware)
	participants.RegisterRoutes(coursePhaseApi, participantsService, promptSDK.AuthenticationMiddleware)
	generator.RegisterRoutes(coursePhaseApi, generatorService, promptSDK.AuthenticationMiddleware)
	privacy.RegisterRoutes(api, privacyService)
	copy.RegisterRoutes(api, copyService)
	coursePhaseDeletion.RegisterRoutes(coursePhaseApi, coursePhaseDeletionService)
	return nil
}

func helloCertificate(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "Hello from certificate service",
	})
}
