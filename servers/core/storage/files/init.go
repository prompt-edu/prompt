package files

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/storage"
	log "github.com/sirupsen/logrus"
)

// Init initializes the application documents file storage module.
func Init(queries db.Queries, conn *pgxpool.Pool) error {
	maxFileSizeMB := int64(50) // Default 50MB

	allowedTypesStr := sdkUtils.GetEnv("ALLOWED_FILE_TYPES", "")
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
		return fmt.Errorf("missing S3 credentials for non-local endpoint: set S3_ACCESS_KEY and S3_SECRET_KEY")
	}

	adapter, err := storage.NewS3Adapter(bucket, region, endpoint, publicEndpoint, accessKey, secretKey, forcePathStyle)
	if err != nil {
		return fmt.Errorf("failed to create S3 adapter: %w", err)
	}

	log.WithFields(log.Fields{
		"bucket":         bucket,
		"region":         region,
		"endpoint":       endpoint,
		"forcePathStyle": forcePathStyle,
	}).Info("Initialized application documents S3 adapter")

	StorageServiceSingleton = NewStorageService(queries, conn, adapter, maxFileSizeMB, allowedTypes)

	log.WithFields(log.Fields{
		"maxFileSizeMB": maxFileSizeMB,
		"allowedTypes":  allowedTypes,
	}).Info("Application documents storage service initialized")

	return nil
}
