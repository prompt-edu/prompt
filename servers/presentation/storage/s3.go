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
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
		// Browsers consume these URLs and cannot send the x-amz-checksum-mode header the SDK would otherwise sign.
		options.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	}))
	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil {
		// Only a genuine "not found" justifies creating it. Reporting a 403 or a timeout
		// as a failed bucket creation sends operators chasing the wrong problem.
		var notFound *types.NotFound
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("check S3 bucket existence: %w", err)
		}
		createInput := &s3.CreateBucketInput{Bucket: aws.String(bucket)}
		if region != "" && region != "us-east-1" {
			createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraint(region),
			}
		}
		if _, createErr := client.CreateBucket(ctx, createInput); createErr != nil {
			return nil, fmt.Errorf("create S3 bucket: %w", createErr)
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
		// Materials are downloads, never documents to render in the browser origin.
		ResponseContentDisposition: aws.String("attachment"),
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

// DeletePrefix lists the objects and deletes them one by one rather than through
// DeleteObjects: the single-object call is the one already proven against the deployed
// S3-compatible store.
func (a *S3Adapter) DeletePrefix(ctx context.Context, prefix string) error {
	if err := ValidateDeletePrefix(prefix); err != nil {
		return fmt.Errorf("delete S3 objects: %w", err)
	}
	pages := s3.NewListObjectsV2Paginator(a.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(a.bucket), Prefix: aws.String(prefix),
	})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list S3 objects under %q: %w", prefix, err)
		}
		for _, object := range page.Contents {
			if err := a.Delete(ctx, aws.ToString(object.Key)); err != nil {
				return err
			}
		}
	}
	return nil
}
