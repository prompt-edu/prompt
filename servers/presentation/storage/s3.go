package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type S3Adapter struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

func NewS3Adapter(ctx context.Context, bucket, region, endpoint, publicEndpoint, accessKey, secretKey string, forcePathStyle bool) (*S3Adapter, error) {
	options := []func(*awsConfig.LoadOptions) error{awsConfig.WithRegion(region)}
	if accessKey != "" && secretKey != "" {
		options = append(options, awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	}
	cfg, err := awsConfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("load S3 configuration: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = forcePathStyle
		if endpoint != "" {
			options.BaseEndpoint = aws.String(endpoint)
		}
	})
	presignEndpoint := endpoint
	if publicEndpoint != "" {
		presignEndpoint = publicEndpoint
	}
	presignClient := s3.NewPresignClient(s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = forcePathStyle
		if presignEndpoint != "" {
			options.BaseEndpoint = aws.String(presignEndpoint)
		}
	}))
	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil {
		if _, createErr := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)}); createErr != nil {
			return nil, fmt.Errorf("ensure S3 bucket: %w", createErr)
		}
	}
	return &S3Adapter{client: client, presignClient: presignClient, bucket: bucket}, nil
}

func (a *S3Adapter) GetUploadURL(ctx context.Context, key, contentType string, ttlSeconds int) (string, error) {
	request, err := a.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucket), Key: aws.String(key), ContentType: aws.String(contentType),
	}, func(options *s3.PresignOptions) { options.Expires = time.Duration(ttlSeconds) * time.Second })
	if err != nil {
		return "", fmt.Errorf("presign upload: %w", err)
	}
	return request.URL, nil
}

func (a *S3Adapter) GetDownloadURL(ctx context.Context, key string, ttlSeconds int) (string, error) {
	request, err := a.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket), Key: aws.String(key),
	}, func(options *s3.PresignOptions) { options.Expires = time.Duration(ttlSeconds) * time.Second })
	if err != nil {
		return "", fmt.Errorf("presign download: %w", err)
	}
	return request.URL, nil
}

func (a *S3Adapter) GetMetadata(ctx context.Context, key string) (Metadata, error) {
	response, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(a.bucket), Key: aws.String(key)})
	if err != nil {
		var apiErr smithy.APIError
		var responseErr smithyhttpResponseError
		if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey") {
			return Metadata{}, ErrObjectNotFound
		}
		if errors.As(err, &responseErr) && responseErr.HTTPStatusCode() == http.StatusNotFound {
			return Metadata{}, ErrObjectNotFound
		}
		return Metadata{}, fmt.Errorf("read S3 metadata: %w", err)
	}
	metadata := Metadata{}
	if response.ContentLength != nil {
		metadata.Size = *response.ContentLength
	}
	if response.ContentType != nil {
		metadata.ContentType = *response.ContentType
	}
	return metadata, nil
}

// smithyhttpResponseError is the subset exposed by transport response errors.
type smithyhttpResponseError interface {
	error
	HTTPStatusCode() int
}

func (a *S3Adapter) Delete(ctx context.Context, key string) error {
	if _, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(a.bucket), Key: aws.String(key)}); err != nil {
		return fmt.Errorf("delete S3 object: %w", err)
	}
	return nil
}
