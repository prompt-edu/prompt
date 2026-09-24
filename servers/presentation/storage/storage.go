package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
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
	// DeletePrefix removes every object whose key starts with prefix, so a retry after a
	// failure still reaches the objects left behind.
	DeletePrefix(ctx context.Context, prefix string) error
}

// CoursePhasePrefix is the key prefix every material of the course phase is stored under.
// The trailing slash keeps the prefix from matching a longer key segment.
func CoursePhasePrefix(coursePhaseID uuid.UUID) string {
	return "presentations/" + coursePhaseID.String() + "/"
}
