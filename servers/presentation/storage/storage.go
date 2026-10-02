package storage

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrObjectNotFound = errors.New("storage object not found")
	ErrInvalidPrefix  = errors.New("invalid storage prefix")
)

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
	// failure still reaches the objects left behind. It refuses a prefix that
	// ValidateDeletePrefix rejects.
	DeletePrefix(ctx context.Context, prefix string) error
}

// CoursePhasePrefix is the key prefix every material of the course phase is stored under.
// The trailing slash keeps the prefix from matching a longer key segment.
func CoursePhasePrefix(coursePhaseID uuid.UUID) string {
	return "presentations/" + coursePhaseID.String() + "/"
}

// ValidateDeletePrefix accepts only a prefix naming a nested directory, such as a
// CoursePhasePrefix. A prefix without the trailing slash would also match sibling keys
// that merely start the same way, and a top-level one like "presentations/" would wipe
// every course phase at once.
func ValidateDeletePrefix(prefix string) error {
	directory, hasTrailingSlash := strings.CutSuffix(prefix, "/")
	segments := strings.Split(directory, "/")
	if !hasTrailingSlash || len(segments) < 2 || slices.Contains(segments, "") {
		return fmt.Errorf("%w %q: expected a nested directory ending in a slash", ErrInvalidPrefix, prefix)
	}
	return nil
}
