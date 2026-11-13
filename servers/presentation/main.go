package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/presentation/bodyLimit"
	"github.com/prompt-edu/prompt/servers/presentation/coursePhaseDeletion"
	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/presentation/presentation"
	"github.com/prompt-edu/prompt/servers/presentation/storage"
	log "github.com/sirupsen/logrus"
)

const (
	materialReaperInterval  = time.Hour
	materialReaperBatchSize = 100
)

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

// Deliberately separate from core's ALLOWED_FILE_TYPES: presentation materials are slide
// decks, and widening the shared variable would also widen core's document uploads. The
// default is the union of the material catalog, so no upload a lecturer can require is
// blocked by the deployment-wide gate.
func allowedMaterialTypes() []string {
	raw := promptSDK.GetEnv(
		"PRESENTATION_ALLOWED_FILE_TYPES",
		strings.Join(presentation.DefaultAllowedMaterialTypes(), ","),
	)
	types := make([]string, 0)
	for _, value := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			types = append(types, trimmed)
		}
	}
	return types
}

// Reclaims uploads that were presigned but never completed, freeing both the row and the
// stored object. Runs in-process; the claim query is safe to run from several replicas.
func startMaterialReaper(ctx context.Context, service *presentation.Service) {
	ticker := time.NewTicker(materialReaperInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reclaimed, err := service.ReclaimExpiredMaterials(ctx, materialReaperBatchSize)
				if err != nil {
					log.WithError(err).Warn("Could not reclaim expired presentation materials")
					continue
				}
				if reclaimed > 0 {
					log.WithField("count", reclaimed).Info("Reclaimed expired presentation materials")
				}
			}
		}
	}()
}

func main() {
	err := promptSDK.Bootstrap(promptSDK.ServiceOptions{
		ServiceName:    "presentation",
		BasePath:       "/presentation/api",
		DBEnvPrefix:    "PRESENTATION",
		DefaultDBPort:  "5440",
		SentryDSNEnv:   "SENTRY_DSN_PRESENTATION",
		DefaultAddress: "localhost:8089",
		Capabilities: map[string]bool{
			promptTypes.CapabilityPhaseCopy:       true,
			promptTypes.CapabilityPhaseConfig:     true,
			promptTypes.CapabilityPhaseDeletion:   true,
			promptTypes.CapabilityPrivacyExport:   true,
			promptTypes.CapabilityPrivacyDeletion: true,
		},
		RegisterRoutes: registerRoutes,
	})
	if err != nil {
		log.Fatalf("Presentation server stopped: %v", err)
	}
}

func registerRoutes(api, coursePhaseAPI *gin.RouterGroup, conn *pgxpool.Pool) error {
	ctx := context.Background()
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
		return fmt.Errorf("unable to initialize presentation material storage: %w", err)
	}

	queries := db.New(conn)
	service := presentation.NewService(
		queries,
		conn,
		storageAdapter,
		sdkUtils.GetCoreUrl(),
		intEnv("S3_PRESIGN_UPLOAD_TTL_SECONDS", 900),
		intEnv("S3_PRESIGN_DOWNLOAD_TTL_SECONDS", 900),
		int64(intEnv("MAX_FILE_UPLOAD_SIZE_MB", 50))*1024*1024,
		allowedMaterialTypes(),
	)
	startMaterialReaper(ctx, service)

	registerServiceRoutes(api, coursePhaseAPI, service, coursePhaseDeletion.NewCoursePhaseDeletionService(queries, conn, storageAdapter))
	return nil
}

func registerServiceRoutes(api, coursePhaseAPI *gin.RouterGroup, service *presentation.Service, deletionService *coursePhaseDeletion.CoursePhaseDeletionService) {
	// Bootstrap creates coursePhaseAPI before calling back, and Gin snapshots a group's chain when
	// a subgroup is created, so the cap goes on both groups. Routes that need more raise it with
	// bodyLimit.WithLimit.
	api.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))
	coursePhaseAPI.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))

	presentation.RegisterRoutes(coursePhaseAPI, service)
	coursePhaseDeletion.RegisterRoutes(coursePhaseAPI, deletionService)

	promptTypes.RegisterCopyModule(
		api.Group("", audit.Describe(presentation.AuditCopyAction)),
		&presentation.CopyHandler{Service: service},
		promptSDK.PromptAdmin,
		promptSDK.CourseLecturer,
	)
	promptTypes.RegisterPrivacyModule(api, service.PrivacyExportHandler, service.PrivacyDeletionHandler, []string{})
}
