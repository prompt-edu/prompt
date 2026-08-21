package testutils

import (
	"context"
	"sync"

	"github.com/prompt-edu/prompt/servers/presentation/storage"
)

// FakeStorage records what the service asks of object storage, so upload and cleanup
// tests can assert on it without an S3 container.
type FakeStorage struct {
	mu      sync.Mutex
	objects map[string]storage.Metadata
	Deleted []string
}

func NewFakeStorage() *FakeStorage {
	return &FakeStorage{objects: map[string]storage.Metadata{}}
}

// Put simulates the browser's PUT to the presigned URL.
func (f *FakeStorage) Put(key, contentType string, size int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[key] = storage.Metadata{ContentType: contentType, Size: size}
}

func (f *FakeStorage) Has(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.objects[key]
	return ok
}

func (f *FakeStorage) GetUploadURL(_ context.Context, key, _ string, _ int) (string, error) {
	return "https://storage.test/" + key, nil
}

func (f *FakeStorage) GetDownloadURL(_ context.Context, key string, _ int) (string, error) {
	return "https://storage.test/" + key, nil
}

func (f *FakeStorage) GetMetadata(_ context.Context, key string) (storage.Metadata, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	metadata, ok := f.objects[key]
	if !ok {
		return storage.Metadata{}, storage.ErrObjectNotFound
	}
	return metadata, nil
}

func (f *FakeStorage) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.objects, key)
	f.Deleted = append(f.Deleted, key)
	return nil
}
