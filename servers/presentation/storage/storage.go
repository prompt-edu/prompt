package storage

import (
	"context"
	"errors"
)

var ErrObjectNotFound = errors.New("storage object not found")

type Metadata struct {
	ContentType string
	Size        int64
}

type Adapter interface {
	GetUploadURL(ctx context.Context, key, contentType string, ttlSeconds int) (string, error)
	GetDownloadURL(ctx context.Context, key string, ttlSeconds int) (string, error)
	GetMetadata(ctx context.Context, key string) (Metadata, error)
	Delete(ctx context.Context, key string) error
}
