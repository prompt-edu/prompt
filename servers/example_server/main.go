package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/example_server/bodyLimit"
	"github.com/prompt-edu/prompt/servers/example_server/config"
	"github.com/prompt-edu/prompt/servers/example_server/copy"
	db "github.com/prompt-edu/prompt/servers/example_server/db/sqlc"
	"github.com/prompt-edu/prompt/servers/example_server/example"
	log "github.com/sirupsen/logrus"
)

// @title           PROMPT Example API
// @version         1.0
// @description     This is the example server of PROMPT.
// @host            localhost:8086
// @BasePath        /example-service/api
// @externalDocs.description  PROMPT Documentation
// @externalDocs.url          https://prompt-edu.github.io/prompt/
func main() {
	err := promptSDK.Bootstrap(promptSDK.ServiceOptions{
		ServiceName:    "example-service",
		BasePath:       "/example-service/api",
		DBEnvPrefix:    "EXAMPLE_SERVER",
		DefaultDBPort:  "5437",
		SentryDSNEnv:   "SENTRY_DSN_EXAMPLE_SERVER",
		DefaultAddress: "localhost:8086",
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   false,
			promptTypes.CapabilityPrivacyDeletion: false,
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
		},
		RegisterRoutes: registerRoutes,
	})
	if err != nil {
		log.Fatalf("Example server stopped: %v", err)
	}
}

func registerRoutes(api, coursePhaseApi *gin.RouterGroup, conn *pgxpool.Pool) error {
	// Bootstrap creates coursePhaseApi before calling back, and Gin snapshots a group's chain when
	// a subgroup is created, so the cap goes on both groups. Routes that need more raise it with
	// bodyLimit.WithLimit.
	api.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))
	coursePhaseApi.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))

	query := db.New(conn)

	coursePhaseApi.GET("/hello", helloExampleServer)

	configService := config.NewConfigService(*query, conn)
	copyService := copy.NewCopyService(*query, conn)
	exampleService := example.NewExampleService(*query, conn)

	copy.RegisterRoutes(api, copyService)
	config.RegisterRoutes(coursePhaseApi, configService)
	example.RegisterRoutes(coursePhaseApi, exampleService, promptSDK.AuthenticationMiddleware)
	return nil
}

// helloExampleServer godoc
// @Summary Example server health check
// @Description Returns a simple hello message from the example server.
// @Tags example
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/hello [get]
func helloExampleServer(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello from the example service",
	})
}
