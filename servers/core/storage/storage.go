package storage

import (
	"context"
	"io"
)

// FileMetadata represents metadata about an uploaded file
type FileMetadata struct {
	StorageKey  string
	Filename    string
	ContentType string
	Size        int64
}

// UploadResult contains the result of a file upload operation
type UploadResult struct {
	StorageKey string
	PublicURL  string
	Size       int64
}

// StorageAdapter defines the interface for file storage operations
// This allows different storage backends (SeaweedFS, S3, etc.) to be used interchangeably
type StorageAdapter interface {
	// Upload stores a file and returns metadata about the stored file
	// The reader will be consumed entirely, and the method takes ownership of closing it if needed
	Upload(ctx context.Context, storageKey string, contentType string, reader io.Reader) (*UploadResult, error)

	// Download retrieves a file by its storage key
	// The caller is responsible for closing the returned reader
	Download(ctx context.Context, storageKey string) (io.ReadCloser, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, storageKey string) error

	// GetUploadURL returns a (potentially temporary/signed) URL for uploading the file
	GetUploadURL(ctx context.Context, storageKey string, contentType string, ttl int) (string, error)

	// GetURL returns a (potentially temporary/signed) URL for accessing the file
	// If ttl is 0, returns a permanent URL (if supported by the storage backend)
	GetURL(ctx context.Context, storageKey string, ttl int) (string, error)

	// GetMetadata retrieves metadata about a stored file without downloading it
	GetMetadata(ctx context.Context, storageKey string) (*FileMetadata, error)
}
