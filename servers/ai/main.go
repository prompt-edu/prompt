package main

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/ai/calls"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
	"github.com/prompt-edu/prompt/servers/ai/encryption"
	"github.com/prompt-edu/prompt/servers/ai/gateway"
	"github.com/prompt-edu/prompt/servers/ai/privacy"
	"github.com/prompt-edu/prompt/servers/ai/retention"
	log "github.com/sirupsen/logrus"
)

func databaseURL() string {
	value := &url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			promptSDK.GetEnv("DB_USER", "prompt-postgres"),
			promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres"),
		),
		Host: fmt.Sprintf(
			"%s:%s",
			promptSDK.GetEnv("DB_HOST_AI", "localhost"),
			promptSDK.GetEnv("DB_PORT_AI", "5442"),
		),
		Path: "/" + promptSDK.GetEnv("DB_NAME", "prompt"),
	}
	query := value.Query()
	query.Set("sslmode", promptSDK.GetEnv("SSL_MODE", "disable"))
	query.Set("TimeZone", promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin"))
	value.RawQuery = query.Encode()
	return value.String()
}

type config struct {
	providerURL           *url.URL
	allowedModels         []string
	metadataRetentionDays int
	serverVersion         string
	coreURL               string
	clientHost            string
}

func loadConfig() (config, error) {
	providerURL, err := url.Parse(promptSDK.GetEnv("AI_PROVIDER_BASE_URL", ""))
	if err != nil || providerURL.Scheme == "" || providerURL.Host == "" {
		return config{}, fmt.Errorf("AI_PROVIDER_BASE_URL must be an absolute URL such as https://logos.example/v1")
	}
	var allowedModels []string
	for _, model := range strings.Split(promptSDK.GetEnv("AI_ALLOWED_MODELS", ""), ",") {
		if trimmed := strings.TrimSpace(model); trimmed != "" {
			allowedModels = append(allowedModels, trimmed)
		}
	}
	if len(allowedModels) == 0 {
		return config{}, fmt.Errorf("AI_ALLOWED_MODELS must list at least one model")
	}
	metadataRetentionDays, err := strconv.Atoi(promptSDK.GetEnv("AI_AUDIT_METADATA_RETENTION_DAYS", "1825"))
	if err != nil {
		return config{}, fmt.Errorf("AI_AUDIT_METADATA_RETENTION_DAYS must be a number of days")
	}
	if err := retention.Validate(metadataRetentionDays); err != nil {
		return config{}, err
	}
	if err := encryption.ValidateKey(); err != nil {
		return config{}, err
	}
	return config{
		providerURL:           providerURL,
		allowedModels:         allowedModels,
		metadataRetentionDays: metadataRetentionDays,
		serverVersion:         promptSDK.GetEnv("SERVER_IMAGE_TAG", ""),
		coreURL:               sdkUtils.GetCoreUrl(),
		clientHost:            promptSDK.GetEnv("CORE_HOST", "http://localhost:3000"),
	}, nil
}

// No audit middleware: AI calls are audited in this service's own database, never in core's.
func setupRouter(router *gin.Engine, conn *pgxpool.Pool, cfg config) {
	queries := db.New(conn)
	callService := calls.NewService(queries, conn, cfg.providerURL.Host, cfg.serverVersion)
	privacyService := privacy.NewService(queries, conn)

	router.Use(promptSDK.CORSMiddleware(cfg.clientHost))
	api := router.Group("/ai/api")
	coursePhaseAPI := api.Group("/course_phase/:coursePhaseID")
	gateway.RegisterRoutes(coursePhaseAPI, gateway.New(cfg.providerURL, cfg.allowedModels, cfg.coreURL, callService))
	calls.RegisterRoutes(coursePhaseAPI, callService)
	promptTypes.RegisterPrivacyModule(api, privacyService.Export, privacyService.Delete, []string{})
	promptTypes.RegisterInfoEndpoint(api, promptTypes.ServiceInfo{
		ServiceName: "ai",
		Version:     cfg.serverVersion,
		Capabilities: map[string]bool{
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
		},
	}, func() bool {
		pingContext, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		return conn.Ping(pingContext) == nil
	})
}

func main() {
	sentryEnabled := promptSDK.GetEnv("SENTRY_ENABLED", "false") == "true"
	if sentryEnabled {
		_ = sdkUtils.InitSentry(promptSDK.GetEnv("SENTRY_DSN_AI", ""))
		defer sentry.Flush(2 * time.Second)
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Invalid AI server configuration: %v", err)
	}

	connectionURL := databaseURL()
	if err := sdkUtils.RunMigrations(connectionURL, "./db/migration"); err != nil {
		log.Fatalf("Failed to run AI migrations: %v", err)
	}
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, connectionURL)
	if err != nil {
		log.Fatalf("Unable to create AI database pool: %v", err)
	}
	defer conn.Close()

	if err := promptSDK.InitPhaseKeycloak(); err != nil {
		log.Fatalf("Failed to initialize Keycloak: %v", err)
	}
	retention.Start(ctx, db.New(conn), cfg.metadataRetentionDays)

	router := gin.Default()
	if sentryEnabled {
		router.Use(sentrygin.New(sentrygin.Options{}))
	}
	setupRouter(router, conn, cfg)

	address := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8092")
	log.Infof("AI server started on %s", address)
	if err := router.Run(address); err != nil {
		log.Fatalf("AI server stopped: %v", err)
	}
}
