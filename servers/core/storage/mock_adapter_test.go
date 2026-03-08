package storage

import (
	"bytes"
	"context"
	"io"
)

// MockStorageAdapter is a mock implementation of StorageAdapter for testing
type MockStorageAdapter struct {
	UploadFunc      func(ctx context.Context, storageKey string, contentType string, reader io.Reader) (*UploadResult, error)
	DownloadFunc    func(ctx context.Context, storageKey string) (io.ReadCloser, error)
	DeleteFunc      func(ctx context.Context, storageKey string) error
	GetUploadURLFunc func(ctx context.Context, storageKey string, contentType string, ttl int) (string, error)
	GetURLFunc      func(ctx context.Context, storageKey string, ttl int) (string, error)
	GetMetadataFunc func(ctx context.Context, storageKey string) (*FileMetadata, error)
}

func (m *MockStorageAdapter) Upload(ctx context.Context, storageKey string, contentType string, reader io.Reader) (*UploadResult, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, storageKey, contentType, reader)
	}
	// Default mock behavior - read the actual content to get size
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return &UploadResult{
		StorageKey: "mock-storage-key/" + storageKey,
		PublicURL:  "https://mock-storage.example.com/mock-storage-key/" + storageKey,
		Size:       int64(len(content)),
	}, nil
}

func (m *MockStorageAdapter) Download(ctx context.Context, storageKey string) (io.ReadCloser, error) {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, storageKey)
	}
	return io.NopCloser(bytes.NewReader([]byte("mock file content"))), nil
}

func (m *MockStorageAdapter) Delete(ctx context.Context, storageKey string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, storageKey)
	}
	return nil
}

func (m *MockStorageAdapter) GetUploadURL(ctx context.Context, storageKey string, contentType string, ttl int) (string, error) {
	if m.GetUploadURLFunc != nil {
		return m.GetUploadURLFunc(ctx, storageKey, contentType, ttl)
	}
	return "https://mock-s3.example.com/" + storageKey + "?presigned=upload", nil
}

func (m *MockStorageAdapter) GetURL(ctx context.Context, storageKey string, ttl int) (string, error) {
	if m.GetURLFunc != nil {
		return m.GetURLFunc(ctx, storageKey, ttl)
	}
	return "https://mock-s3.example.com/" + storageKey + "?presigned=true", nil
}

func (m *MockStorageAdapter) GetMetadata(ctx context.Context, storageKey string) (*FileMetadata, error) {
	if m.GetMetadataFunc != nil {
		return m.GetMetadataFunc(ctx, storageKey)
	}
	return &FileMetadata{
		StorageKey:  storageKey,
		Size:        1024,
		ContentType: "application/pdf",
		Filename:    "mock-file.pdf",
	}, nil
}
