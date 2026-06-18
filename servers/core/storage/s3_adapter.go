package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	log "github.com/sirupsen/logrus"
)

// S3Adapter implements the StorageAdapter interface using AWS S3 or compatible services
// Compatible with AWS S3, SeaweedFS S3 gateway, MinIO, and other S3-compatible storage
type S3Adapter struct {
	client          *s3.Client
	presignClient   *s3.PresignClient
	bucket          string
	forcePathStyle  bool
	presignDuration time.Duration
	publicEndpoint  string
}

// make StorageAdapter implementation explicit
var _ StorageAdapter = (*S3Adapter)(nil)

// NewS3Adapter creates a new S3 storage adapter
// Works with AWS S3, SeaweedFS S3 gateway, MinIO, and other S3-compatible services
func NewS3Adapter(bucket, region, endpoint, publicEndpoint, accessKey, secretKey string, forcePathStyle bool) (*S3Adapter, error) {
	ctx := context.Background()
	var cfg aws.Config
	var err error
	cfg, err = buildS3Config(ctx, region, accessKey, secretKey)

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = forcePathStyle
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	presignEndpoint := endpoint
	if publicEndpoint != "" {
		presignEndpoint = publicEndpoint
	}

	presignClient := s3.NewPresignClient(s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = forcePathStyle
		if presignEndpoint != "" {
			o.BaseEndpoint = aws.String(presignEndpoint)
		}
	}))

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		var notFoundErr *types.NotFound
		if !errors.As(err, &notFoundErr) {
			return nil, fmt.Errorf("failed to check bucket existence: %w", err)
		}

		log.WithField("bucket", bucket).Info("Bucket does not exist, attempting to create it")
		createInput := &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		}
		if region != "" && region != "us-east-1" {
			createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraint(region),
			}
		}

		if _, createErr := client.CreateBucket(ctx, createInput); createErr != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", createErr)
		}

		log.WithField("bucket", bucket).Info("Bucket created successfully")
	}

	log.WithFields(log.Fields{
		"bucket":         bucket,
		"region":         region,
		"endpoint":       endpoint,
		"forcePathStyle": forcePathStyle,
	}).Info("S3 adapter initialized")

	return &S3Adapter{
		client:          client,
		presignClient:   presignClient,
		bucket:          bucket,
		forcePathStyle:  forcePathStyle,
		presignDuration: time.Minute, // Default presign duration
		publicEndpoint:  publicEndpoint,
	}, nil
}

func buildS3Config(ctx context.Context, region, accessKey, secretKey string) (aws.Config, error) {
	if accessKey == "" || secretKey == "" {
		return config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	return config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
}

// Upload stores a file in S3
func (s *S3Adapter) Upload(ctx context.Context, storageKey string, contentType string, reader io.Reader) (*UploadResult, error) {
	key := storageKey
	if key == "" {
		return nil, fmt.Errorf("storage key cannot be empty")
	}

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		log.WithError(err).WithField("key", key).Error("Failed to upload file to S3")
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Get object size
	head, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	var size int64
	if err == nil && head.ContentLength != nil {
		size = *head.ContentLength
	}

	log.WithFields(log.Fields{
		"key":  key,
		"size": size,
	}).Info("File uploaded to S3 successfully")

	return &UploadResult{
		StorageKey: key,
		PublicURL:  fmt.Sprintf("%s/%s", s.bucket, key), // Note: This is not a real URL, just the key
		Size:       size,
	}, nil
}

// Download retrieves a file from S3
func (s *S3Adapter) Download(ctx context.Context, storageKey string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	})

	if err != nil {
		log.WithError(err).WithField("key", storageKey).Error("Failed to download file from S3")
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return result.Body, nil
}

// Delete removes a file from S3
func (s *S3Adapter) Delete(ctx context.Context, storageKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	})

	if err != nil {
		log.WithError(err).WithField("key", storageKey).Error("Failed to delete file from S3")
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	log.WithField("key", storageKey).Info("File deleted from S3 successfully")
	return nil
}

// GetUploadURL returns a presigned URL for uploading a file
func (s *S3Adapter) GetUploadURL(ctx context.Context, storageKey string, contentType string, ttl int) (string, error) {
	duration := s.presignDuration
	if ttl > 0 {
		duration = time.Duration(ttl) * time.Second
	}

	request, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(storageKey),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		log.WithError(err).WithField("key", storageKey).Error("Failed to generate presigned upload URL")
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return request.URL, nil
}

// GetURL returns a presigned URL for accessing the file
func (s *S3Adapter) GetURL(ctx context.Context, storageKey string, ttl int) (string, error) {
	duration := s.presignDuration
	if ttl > 0 {
		duration = time.Duration(ttl) * time.Second
	}

	request, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})

	if err != nil {
		log.WithError(err).WithField("key", storageKey).Error("Failed to generate presigned URL")
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// GetMetadata retrieves metadata about a file
func (s *S3Adapter) GetMetadata(ctx context.Context, storageKey string) (*FileMetadata, error) {
	head, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	})

	if err != nil {
		log.WithError(err).WithField("key", storageKey).Error("Failed to get file metadata from S3")
		return nil, fmt.Errorf("failed to get metadata from S3: %w", err)
	}

	contentType := ""
	if head.ContentType != nil {
		contentType = *head.ContentType
	}

	size := int64(0)
	if head.ContentLength != nil {
		size = *head.ContentLength
	}

	return &FileMetadata{
		StorageKey:  storageKey,
		Filename:    storageKey, // S3 key is the filename
		ContentType: contentType,
		Size:        size,
	}, nil
}
