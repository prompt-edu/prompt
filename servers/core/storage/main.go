package storage

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// InitStorageModule initializes the storage module with service only (no public routes)
func InitStorageModule(queries db.Queries, conn *pgxpool.Pool) error {
	// Get storage configuration from environment
	maxFileSizeMB := int64(50) // Default 50MB

	// Parse allowed file types
	allowedTypesStr := sdkUtils.GetEnv("ALLOWED_FILE_TYPES", "")
	var allowedTypes []string
	if allowedTypesStr != "" {
		allowedTypes = strings.Split(allowedTypesStr, ",")
		for i := range allowedTypes {
			allowedTypes[i] = strings.TrimSpace(allowedTypes[i])
		}
	}

	// S3 configuration (works with AWS S3, SeaweedFS S3 gateway, MinIO, etc.)
	bucket := sdkUtils.GetEnv("S3_BUCKET", "prompt-files")
	region := sdkUtils.GetEnv("S3_REGION", "us-east-1")
	endpoint := sdkUtils.GetEnv("S3_ENDPOINT", "http://localhost:8334") // Empty for AWS S3, set for SeaweedFS/MinIO
	publicEndpoint := sdkUtils.GetEnv("S3_PUBLIC_ENDPOINT", "")
	accessKey := sdkUtils.GetEnv("S3_ACCESS_KEY", "")
	secretKey := sdkUtils.GetEnv("S3_SECRET_KEY", "")
	forcePathStyle := sdkUtils.GetEnv("S3_FORCE_PATH_STYLE", "true") == "true" // Required for SeaweedFS/MinIO

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

	adapter, err := NewS3Adapter(bucket, region, endpoint, publicEndpoint, accessKey, secretKey, forcePathStyle)
	if err != nil {
		return fmt.Errorf("failed to create S3 adapter: %w", err)
	}

	log.WithFields(log.Fields{
		"bucket":         bucket,
		"region":         region,
		"endpoint":       endpoint,
		"forcePathStyle": forcePathStyle,
	}).Info("Initialized S3-compatible storage adapter")

	// Create storage service singleton
	StorageServiceSingleton = NewStorageService(queries, conn, adapter, maxFileSizeMB, allowedTypes)

	log.WithFields(log.Fields{
		"maxFileSizeMB": maxFileSizeMB,
		"allowedTypes":  allowedTypes,
	}).Info("Storage service initialized")

	return nil
}
