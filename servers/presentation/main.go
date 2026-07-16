package main

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/presentation/presentation"
	"github.com/prompt-edu/prompt/servers/presentation/storage"
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
			promptSDK.GetEnv("DB_HOST_PRESENTATION", "localhost"),
			promptSDK.GetEnv("DB_PORT_PRESENTATION", "5440"),
		),
		Path: "/" + promptSDK.GetEnv("DB_NAME", "prompt"),
	}
	query := value.Query()
	query.Set("sslmode", promptSDK.GetEnv("SSL_MODE", "disable"))
	query.Set("TimeZone", promptSDK.GetEnv("DB_TIMEZONE", "Europe/Berlin"))
	value.RawQuery = query.Encode()
	return value.String()
}

func runMigrations(connectionURL string) {
	command := exec.Command("migrate", "-path", "./db/migration", "-database", connectionURL, "up")
	output, err := command.CombinedOutput()
	sanitized := strings.ReplaceAll(
		string(output),
		promptSDK.GetEnv("DB_PASSWORD", "prompt-postgres"),
		"***",
	)
	if err != nil {
		log.Fatalf("Failed to run presentation migrations: %v\n%s", err, sanitized)
	}
	if sanitized != "" {
		log.Debug(sanitized)
	}
}

func initAuthentication() {
	keycloakURL := promptSDK.GetEnv("KEYCLOAK_HOST", "http://localhost:8081")
	if !strings.HasPrefix(keycloakURL, "http") {
		keycloakURL = "https://" + keycloakURL
	}
	if err := promptSDK.InitAuthenticationMiddleware(
		keycloakURL,
		promptSDK.GetEnv("KEYCLOAK_REALM_NAME", "prompt"),
		sdkUtils.GetCoreUrl(),
	); err != nil {
		log.Fatalf("Failed to initialize Keycloak: %v", err)
	}
}

func intEnv(key string, defaultValue int) int {
	value, err := strconv.Atoi(promptSDK.GetEnv(key, strconv.Itoa(defaultValue)))
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

func boolEnv(key string, defaultValue bool) bool {
	value, err := strconv.ParseBool(promptSDK.GetEnv(key, strconv.FormatBool(defaultValue)))
	if err != nil {
		return defaultValue
	}
	return value
}

func main() {
	connectionURL := databaseURL()
	runMigrations(connectionURL)
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, connectionURL)
	if err != nil {
		log.Fatalf("Unable to create presentation database pool: %v", err)
	}
	defer conn.Close()

	storageAdapter, err := storage.NewS3Adapter(
		ctx,
		promptSDK.GetEnv("S3_BUCKET", "prompt"),
		promptSDK.GetEnv("S3_REGION", "us-east-1"),
		promptSDK.GetEnv("S3_ENDPOINT", ""),
		promptSDK.GetEnv("S3_PUBLIC_ENDPOINT", ""),
		promptSDK.GetEnv("S3_ACCESS_KEY", ""),
		promptSDK.GetEnv("S3_SECRET_KEY", ""),
		boolEnv("S3_FORCE_PATH_STYLE", true),
	)
	if err != nil {
		log.Fatalf("Unable to initialize presentation material storage: %v", err)
	}

	initAuthentication()
	queries := db.New(conn)
	service := presentation.NewService(
		queries,
		conn,
		storageAdapter,
		sdkUtils.GetCoreUrl(),
		intEnv("S3_PRESIGN_UPLOAD_TTL_SECONDS", 900),
		intEnv("S3_PRESIGN_DOWNLOAD_TTL_SECONDS", 900),
		int64(intEnv("MAX_FILE_UPLOAD_SIZE_MB", 50))*1024*1024,
	)

	router := gin.Default()
	router.Use(promptSDK.CORSMiddleware(promptSDK.GetEnv("CORE_HOST", "http://localhost:3000")))
	api := router.Group("/presentation/api")
	coursePhaseAPI := api.Group("/course_phase/:coursePhaseID")
	presentation.RegisterRoutes(coursePhaseAPI, service)

	promptTypes.RegisterCopyEndpoint(
		api,
		promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer),
		&presentation.CopyHandler{Service: service},
	)
	promptTypes.RegisterPrivacyDataExportEndpoint(api, service.PrivacyExportHandler, []string{})
	promptTypes.RegisterPrivacyDataDeletionEndpoint(api, service.PrivacyDeletionHandler)
	promptTypes.RegisterInfoEndpoint(api, promptTypes.ServiceInfo{
		ServiceName: "presentation",
		Version:     promptSDK.GetEnv("SERVER_IMAGE_TAG", ""),
		Capabilities: map[string]bool{
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
		},
	}, func() bool {
		pingContext, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		return conn.Ping(pingContext) == nil
	})

	address := promptSDK.GetEnv("SERVER_ADDRESS", "localhost:8089")
	log.Infof("Presentation server started on %s", address)
	if err := router.Run(address); err != nil {
		log.Fatalf("Presentation server stopped: %v", err)
	}
}
