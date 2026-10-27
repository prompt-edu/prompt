package files

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/storage"
	log "github.com/sirupsen/logrus"
)

// defaultAllowedFileTypes is a safe allowlist used when ALLOWED_FILE_TYPES is unset, so a
// deployment never silently falls back to accepting every content type. It matches the
// documented default in .env.template and deliberately excludes types that render as
// active content in the browser (e.g. text/html, image/svg+xml).
const defaultAllowedFileTypes = "application/pdf,image/jpeg,image/png,image/gif,application/zip,text/plain"

// NewStorageServiceFromEnv builds the application documents file storage service from the
// S3 environment configuration.
func NewStorageServiceFromEnv(queries db.Queries, conn *pgxpool.Pool) (*StorageService, error) {
	maxFileSizeMB := int64(50) // Default 50MB
	if v, err := strconv.ParseInt(sdkUtils.GetEnv("MAX_FILE_UPLOAD_SIZE_MB", "50"), 10, 64); err == nil && v > 0 {
		maxFileSizeMB = v
	}

	allowedTypesStr := sdkUtils.GetEnv("ALLOWED_FILE_TYPES", defaultAllowedFileTypes)
	var allowedTypes []string
	if allowedTypesStr != "" {
		allowedTypes = strings.Split(allowedTypesStr, ",")
		for i := range allowedTypes {
			allowedTypes[i] = strings.TrimSpace(allowedTypes[i])
		}
	}

	bucket := sdkUtils.GetEnv("S3_BUCKET", "prompt-files")
	region := sdkUtils.GetEnv("S3_REGION", "us-east-1")
	endpoint := sdkUtils.GetEnv("S3_ENDPOINT", "http://localhost:8334")
	publicEndpoint := sdkUtils.GetEnv("S3_PUBLIC_ENDPOINT", "")
	accessKey := sdkUtils.GetEnv("S3_ACCESS_KEY", "")
	secretKey := sdkUtils.GetEnv("S3_SECRET_KEY", "")
	forcePathStyle := sdkUtils.GetEnv("S3_FORCE_PATH_STYLE", "true") == "true"

	lowerEndpoint := strings.ToLower(endpoint)
	isLocalEndpoint := strings.Contains(lowerEndpoint, "localhost") || strings.Contains(lowerEndpoint, "127.0.0.1")
	if isLocalEndpoint {
		if accessKey == "" {
			accessKey = "admin"
		}
		if secretKey == "" {
			secretKey = "admin123"
		}
	} else if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("missing S3 credentials for non-local endpoint: set S3_ACCESS_KEY and S3_SECRET_KEY")
	}

	adapter, err := storage.NewS3Adapter(bucket, region, endpoint, publicEndpoint, accessKey, secretKey, forcePathStyle)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 adapter: %w", err)
	}

	log.WithFields(log.Fields{
		"bucket":         bucket,
		"region":         region,
		"endpoint":       endpoint,
		"forcePathStyle": forcePathStyle,
	}).Info("Initialized application documents S3 adapter")

	service := NewStorageService(queries, conn, adapter, maxFileSizeMB, allowedTypes)

	log.WithFields(log.Fields{
		"maxFileSizeMB": maxFileSizeMB,
		"allowedTypes":  allowedTypes,
	}).Info("Application documents storage service initialized")

	return service, nil
}
