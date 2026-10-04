package privacyexport

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/storage"
	log "github.com/sirupsen/logrus"
)

const uploadPresignTTL = 15 * time.Minute
const downloadPresignTTL = 5 * time.Minute
const objectKeyFormat = "%s/prompt_%s_%s.zip" // {exportRequestID}/prompt_{exportRequestID[:6]}_{serviceName}.zip

// ExportStorage stores the GDPR export ZIPs of the privacy module.
type ExportStorage struct {
	adapter storage.StorageAdapter
}

// NewExportStorage wraps an already built storage adapter.
func NewExportStorage(adapter storage.StorageAdapter) *ExportStorage {
	return &ExportStorage{adapter: adapter}
}

// NewExportStorageFromEnv builds the privacy export storage adapter.
// It reuses the same S3 credentials and endpoint as the main storage module
// but targets a dedicated bucket (S3_PRIVACY_EXPORT_BUCKET, default "prompt-privacy-exports").
func NewExportStorageFromEnv() (*ExportStorage, error) {
	bucket := sdkUtils.GetEnv("S3_PRIVACY_EXPORT_BUCKET", "prompt-privacy-exports")
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

	s3Adapter, err := storage.NewS3Adapter(bucket, region, endpoint, publicEndpoint, accessKey, secretKey, forcePathStyle)
	if err != nil {
		return nil, fmt.Errorf("failed to create privacy export S3 adapter: %w", err)
	}

	log.WithFields(log.Fields{
		"bucket":   bucket,
		"endpoint": endpoint,
		"uploadTTL":   uploadPresignTTL,
		"downloadTTL": downloadPresignTTL,
	}).Info("Privacy export storage adapter initialized")

	return NewExportStorage(s3Adapter), nil
}

func MakeObjectURL(exportRequestID string, serviceName string) string {
	return fmt.Sprintf(objectKeyFormat, exportRequestID, exportRequestID[:6], serviceName)
}

// GetUploadURL generates a presigned S3 PUT URL for a microservice to upload
// its GDPR export ZIP. The URL is valid for 15 minutes.
func (s *ExportStorage) GetUploadURL(ctx context.Context, exportRequestID, serviceName string) (string, error) {
	if s.adapter == nil {
		return "", fmt.Errorf("privacy export storage not initialized")
	}

	key := MakeObjectURL(exportRequestID, serviceName)
	return s.adapter.GetUploadURL(ctx, key, "application/zip", int(uploadPresignTTL.Seconds()))
}

// GetDownloadURL generates a presigned S3 GET URL for downloading a GDPR export ZIP.
// The URL is valid for 5 minutes.
func (s *ExportStorage) GetDownloadURL(ctx context.Context, objectKey string) (string, error) {
	if s.adapter == nil {
		return "", fmt.Errorf("privacy export storage not initialized")
	}

	return s.adapter.GetURL(ctx, objectKey, int(downloadPresignTTL.Seconds()))
}

// DeleteFile removes a GDPR export ZIP from storage.
func (s *ExportStorage) DeleteFile(ctx context.Context, objectKey string) error {
	if s.adapter == nil {
		return fmt.Errorf("privacy export storage not initialized")
	}
	return s.adapter.Delete(ctx, objectKey)
}

// GetFileSize returns the size of the stored object in bytes.
func (s *ExportStorage) GetFileSize(ctx context.Context, objectKey string) (int64, error) {
	if s.adapter == nil {
		return 0, fmt.Errorf("privacy export storage not initialized")
	}

	meta, err := s.adapter.GetMetadata(ctx, objectKey)
	if err != nil {
		return 0, err
	}
	return meta.Size, nil
}
