package files

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StorageServiceTestSuite struct {
	suite.Suite
	ctx         context.Context
	cleanup     func()
	service     *StorageService
	mockAdapter storage.StorageAdapter
	testUserID  string
}

func (suite *StorageServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up test database
	testDB, cleanup, err := testutils.SetupTestDBWithMigrations(suite.ctx, "../../db/migration", func(pool *pgxpool.Pool) *db.Queries {
		return db.New(pool)
	}, "../../database_dumps/storage_test.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	// Test data IDs from SQL dump
	suite.testUserID = "11111111-1111-1111-1111-111111111111"

	// Create mock adapter
	suite.mockAdapter = &storage.MockStorageAdapter{}

	// Initialize service with mock adapter
	suite.service = NewStorageService(
		*testDB.Queries,
		testDB.Conn,
		suite.mockAdapter,
		50, // 50 MB max file size
		[]string{"application/pdf", "image/jpeg", "image/png"},
	)

}

func (suite *StorageServiceTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *StorageServiceTestSuite) TestUploadFile_Success() {
	// Create a mock multipart file
	fileContent := []byte("test PDF file content")
	fileHeader := suite.createMultipartFileHeader("test-document.pdf", "application/pdf", fileContent)

	req := FileUploadRequest{
		File:           fileHeader,
		UploaderUserID: suite.testUserID,
		UploaderEmail:  "test.user@tum.de",

		Description: "Test file upload",
	}

	result, err := suite.service.UploadFile(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.NotEqual(suite.T(), uuid.Nil, result.ID)
	assert.Equal(suite.T(), "test-document.pdf", result.OriginalFilename)
	assert.Equal(suite.T(), "application/pdf", result.ContentType)
	assert.Equal(suite.T(), int64(len(fileContent)), result.SizeBytes)
}

func (suite *StorageServiceTestSuite) TestUploadFile_FileTooLarge() {
	// Create file larger than max size (50MB)
	largeContent := make([]byte, 51*1024*1024)
	fileHeader := suite.createMultipartFileHeader("large-file.pdf", "application/pdf", largeContent)

	req := FileUploadRequest{
		File:           fileHeader,
		UploaderUserID: suite.testUserID,
		UploaderEmail:  "test.user@tum.de",
	}

	_, err := suite.service.UploadFile(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "file size")
	assert.Contains(suite.T(), err.Error(), "exceeds")
}

func (suite *StorageServiceTestSuite) TestUploadFile_InvalidFileType() {
	fileContent := []byte("malicious executable content")
	fileHeader := suite.createMultipartFileHeader("malicious.exe", "application/x-msdownload", fileContent)

	req := FileUploadRequest{
		File:           fileHeader,
		UploaderUserID: suite.testUserID,
		UploaderEmail:  "test.user@tum.de",
	}

	_, err := suite.service.UploadFile(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not allowed")
}

func (suite *StorageServiceTestSuite) TestGetFileByID_Success() {
	// First upload a file
	fileContent := []byte("test content for retrieval")
	fileHeader := suite.createMultipartFileHeader("retrieval-test.pdf", "application/pdf", fileContent)

	uploadReq := FileUploadRequest{
		File:           fileHeader,
		UploaderUserID: suite.testUserID,
		UploaderEmail:  "test.user@tum.de",
	}

	uploadResult, err := suite.service.UploadFile(suite.ctx, uploadReq)
	assert.NoError(suite.T(), err)

	// Retrieve it
	file, err := suite.service.GetFileByID(suite.ctx, uploadResult.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), file)
	assert.Equal(suite.T(), uploadResult.ID, file.ID)
	assert.Equal(suite.T(), "retrieval-test.pdf", file.OriginalFilename)
}

// TestDownloadURLsPinContentDisposition verifies file downloads are forced as an
// attachment (so an uploaded HTML/SVG cannot render inline on the storage origin),
// while the inline URL used for profile pictures is served inline. Regression test
// for GHSA-v29x-4vmx-g5g7.
func (suite *StorageServiceTestSuite) TestDownloadURLsPinContentDisposition() {
	mock := suite.mockAdapter.(*storage.MockStorageAdapter)
	var gotDisposition string
	mock.GetURLFunc = func(_ context.Context, storageKey string, _ int, contentDisposition string) (string, error) {
		gotDisposition = contentDisposition
		return "https://example.com/" + storageKey, nil
	}
	defer func() { mock.GetURLFunc = nil }()

	fileHeader := suite.createMultipartFileHeader("disposition.pdf", "application/pdf", []byte("content"))
	uploadResult, err := suite.service.UploadFile(suite.ctx, FileUploadRequest{
		File:           fileHeader,
		UploaderUserID: suite.testUserID,
		UploaderEmail:  "test.user@tum.de",
	})
	require.NoError(suite.T(), err)

	_, err = suite.service.GetFileByID(suite.ctx, uploadResult.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "attachment; filename=disposition.pdf", gotDisposition, "file downloads must be forced as an attachment under their original name")

	_, err = suite.service.GetInlineURL(suite.ctx, "profile-picture-key", 60)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "inline", gotDisposition, "the inline URL must be served inline")
}

func (suite *StorageServiceTestSuite) TestAttachmentDispositionKeepsOriginalFilename() {
	assert.Equal(suite.T(), `attachment; filename="my cv.pdf"`, attachmentDisposition("my cv.pdf"))
	assert.Equal(suite.T(), "attachment; filename*=utf-8''Lebenslauf%20M%C3%BCller.pdf", attachmentDisposition("Lebenslauf Müller.pdf"))
	assert.Equal(suite.T(), "attachment", attachmentDisposition(""))
}

func (suite *StorageServiceTestSuite) TestGetFileByID_NotFound() {
	nonExistentID := uuid.New()

	file, err := suite.service.GetFileByID(suite.ctx, nonExistentID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), file)
}

func (suite *StorageServiceTestSuite) TestDownloadFile_Success() {
	// Upload a file first
	fileContent := []byte("content to download")
	fileHeader := suite.createMultipartFileHeader("download-test.pdf", "application/pdf", fileContent)

	uploadReq := FileUploadRequest{
		File:           fileHeader,
		UploaderUserID: suite.testUserID,
		UploaderEmail:  "test.user@tum.de",
	}

	uploadResult, err := suite.service.UploadFile(suite.ctx, uploadReq)
	assert.NoError(suite.T(), err)

	// Download it
	reader, filename, err := suite.service.DownloadFile(suite.ctx, uploadResult.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), reader)
	assert.Equal(suite.T(), "download-test.pdf", filename)
	defer func() {
		assert.NoError(suite.T(), reader.Close())
	}()

	// Verify content
	downloadedContent, _ := io.ReadAll(reader)
	assert.NotEmpty(suite.T(), downloadedContent)
}

func (suite *StorageServiceTestSuite) TestDeleteFile_SoftDelete() {
	// Upload a file first
	fileContent := []byte("content to delete")
	fileHeader := suite.createMultipartFileHeader("delete-test.pdf", "application/pdf", fileContent)

	uploadReq := FileUploadRequest{
		File:           fileHeader,
		UploaderUserID: suite.testUserID,
		UploaderEmail:  "test.user@tum.de",
	}

	uploadResult, err := suite.service.UploadFile(suite.ctx, uploadReq)
	assert.NoError(suite.T(), err)

	// Soft delete it
	err = suite.service.DeleteFile(suite.ctx, uploadResult.ID, false)
	assert.NoError(suite.T(), err)

	// Verify file is no longer retrievable
	file, err := suite.service.GetFileByID(suite.ctx, uploadResult.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), file)
}

func (suite *StorageServiceTestSuite) TestDeleteFile_HardDelete() {
	fileContent := []byte("content to hard delete")
	fileHeader := suite.createMultipartFileHeader("hard-delete.pdf", "application/pdf", fileContent)

	uploadResult, err := suite.service.UploadFile(suite.ctx, FileUploadRequest{
		File:           fileHeader,
		UploaderUserID: suite.testUserID,
		UploaderEmail:  "test.user@tum.de",
	})
	assert.NoError(suite.T(), err)

	err = suite.service.DeleteFile(suite.ctx, uploadResult.ID, true)
	assert.NoError(suite.T(), err)

	file, err := suite.service.GetFileByID(suite.ctx, uploadResult.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), file)
}

func (suite *StorageServiceTestSuite) TestDeleteFile_HardDelete_StorageFailureReturnsError() {
	fileContent := []byte("content with failing blob delete")
	fileHeader := suite.createMultipartFileHeader("orphan-blob.pdf", "application/pdf", fileContent)

	uploadResult, err := suite.service.UploadFile(suite.ctx, FileUploadRequest{
		File:           fileHeader,
		UploaderUserID: suite.testUserID,
		UploaderEmail:  "test.user@tum.de",
	})
	assert.NoError(suite.T(), err)

	// Make the storage backend deletion fail; the DB record is removed first, so
	// this leaves an orphaned blob. DeleteFile must surface that, not swallow it.
	// The mock adapter is shared across the whole suite (SetupSuite, not
	// SetupTest), so reset the override afterwards to avoid leaking the failure
	// into later tests.
	mock := suite.mockAdapter.(*storage.MockStorageAdapter)
	mock.DeleteFunc = func(ctx context.Context, storageKey string) error {
		return fmt.Errorf("simulated storage backend failure")
	}
	defer func() { mock.DeleteFunc = nil }()

	err = suite.service.DeleteFile(suite.ctx, uploadResult.ID, true)
	assert.Error(suite.T(), err)

	// The record is gone even though the blob deletion failed (the orphan case).
	file, err := suite.service.GetFileByID(suite.ctx, uploadResult.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), file)
}

func (suite *StorageServiceTestSuite) TestGetFilesByUploader() {
	// Upload multiple files by same user
	for i := 0; i < 3; i++ {
		fileContent := []byte("test content")
		fileHeader := suite.createMultipartFileHeader("user-file.pdf", "application/pdf", fileContent)

		req := FileUploadRequest{
			File:           fileHeader,
			UploaderUserID: suite.testUserID,
			UploaderEmail:  "test.user@tum.de",
		}

		_, err := suite.service.UploadFile(suite.ctx, req)
		assert.NoError(suite.T(), err)
	}

	// List files by uploader
	files, err := suite.service.GetFilesByUploader(suite.ctx, suite.testUserID, 10, 0)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(files), 3)
}

func (suite *StorageServiceTestSuite) TestGetFilesByCoursePhaseID() {
	coursePhaseID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	// Upload files for course phase
	for i := 0; i < 2; i++ {
		fileContent := []byte("course phase content")
		fileHeader := suite.createMultipartFileHeader("course-file.pdf", "application/pdf", fileContent)

		req := FileUploadRequest{
			File:           fileHeader,
			UploaderUserID: suite.testUserID,
			UploaderEmail:  "test.user@tum.de",
			CoursePhaseID:  &coursePhaseID,
		}

		_, err := suite.service.UploadFile(suite.ctx, req)
		assert.NoError(suite.T(), err)
	}

	// List files
	files, err := suite.service.GetFilesByCoursePhaseID(suite.ctx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(files), 2)
}

func (suite *StorageServiceTestSuite) TestPresignUpload_MissingFilename_MapsToInvalidInput() {
	_, err := suite.service.PresignUpload(suite.ctx, PresignUploadRequest{
		ContentType: "application/pdf",
	})
	assert.ErrorIs(suite.T(), err, ErrInvalidInput)
}

func (suite *StorageServiceTestSuite) TestCreateFileFromStorageKey_MissingKey_MapsToInvalidInput() {
	_, err := suite.service.CreateFileFromStorageKey(suite.ctx, CreateFileFromStorageKeyRequest{}, suite.testUserID, "")
	assert.ErrorIs(suite.T(), err, ErrInvalidInput)
}

func (suite *StorageServiceTestSuite) TestCreateFileFromStorageKey_ObjectNotFound_MapsToNotFound() {
	mock := suite.mockAdapter.(*storage.MockStorageAdapter)
	old := mock.GetMetadataFunc
	mock.GetMetadataFunc = func(ctx context.Context, storageKey string) (*storage.FileMetadata, error) {
		return nil, storage.ErrObjectNotFound
	}
	defer func() { mock.GetMetadataFunc = old }()

	coursePhaseID := uuid.New()
	storageKey := "course-phase/" + coursePhaseID.String() + "/" + uuid.New().String() + "-never-uploaded.pdf"

	_, err := suite.service.CreateFileFromStorageKey(suite.ctx, CreateFileFromStorageKeyRequest{
		StorageKey:    storageKey,
		CoursePhaseID: &coursePhaseID,
		ContentType:   "application/pdf",
	}, suite.testUserID, "")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

func (suite *StorageServiceTestSuite) TestCreateFileFromStorageKey_DotDotKey_MapsToInvalidInput() {
	coursePhaseID := uuid.New()
	storageKey := "course-phase/" + coursePhaseID.String() + "/../course-phase/" + uuid.New().String() + "/cv.pdf"

	_, err := suite.service.CreateFileFromStorageKey(suite.ctx, CreateFileFromStorageKeyRequest{
		StorageKey:    storageKey,
		CoursePhaseID: &coursePhaseID,
		ContentType:   "application/pdf",
	}, suite.testUserID, "")
	assert.ErrorIs(suite.T(), err, ErrInvalidInput)
}

// Helper method to create a multipart file header for testing
func (suite *StorageServiceTestSuite) createMultipartFileHeader(filename, contentType string, content []byte) *multipart.FileHeader {
	// Create a buffer to write multipart data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file
	part, err := writer.CreateFormFile("file", filename)
	assert.NoError(suite.T(), err)
	_, err = part.Write(content)
	assert.NoError(suite.T(), err)
	err = writer.Close()
	assert.NoError(suite.T(), err)

	// Parse the multipart form
	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(10 << 20) // 10 MB max memory
	assert.NoError(suite.T(), err)

	// Get the file header
	fileHeaders := form.File["file"]
	if len(fileHeaders) > 0 {
		fileHeader := fileHeaders[0]
		// Set content type in header
		fileHeader.Header.Set("Content-Type", contentType)
		return fileHeader
	}

	return nil
}

func (suite *StorageServiceTestSuite) pendingUploadExists(storageKey string) bool {
	var exists bool
	err := suite.service.conn.QueryRow(suite.ctx, "SELECT EXISTS (SELECT 1 FROM pending_upload WHERE storage_key = $1)", storageKey).Scan(&exists)
	require.NoError(suite.T(), err)
	return exists
}

// TestPresignUpload_SignsSizeAndTracksPendingUpload verifies a declared size is bound into the
// upload URL and every handed-out URL is tracked. Regression test for GHSA-c898-5f75-f8qp.
func (suite *StorageServiceTestSuite) TestPresignUpload_SignsSizeAndTracksPendingUpload() {
	mock := suite.mockAdapter.(*storage.MockStorageAdapter)
	var gotContentLength int64
	mock.GetUploadURLFunc = func(_ context.Context, storageKey string, _ string, contentLength int64, _ int) (string, error) {
		gotContentLength = contentLength
		return "https://example.com/" + storageKey, nil
	}
	defer func() { mock.GetUploadURLFunc = nil }()

	coursePhaseID := uuid.New()
	response, err := suite.service.PresignUpload(suite.ctx, PresignUploadRequest{
		Filename:      "cv.pdf",
		ContentType:   "application/pdf",
		SizeBytes:     2048,
		CoursePhaseID: &coursePhaseID,
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2048), gotContentLength)
	assert.True(suite.T(), suite.pendingUploadExists(response.StorageKey))
}

func (suite *StorageServiceTestSuite) TestPresignUpload_RejectsDeclaredSizeAboveLimit() {
	coursePhaseID := uuid.New()
	_, err := suite.service.PresignUpload(suite.ctx, PresignUploadRequest{
		Filename:      "huge.pdf",
		ContentType:   "application/pdf",
		SizeBytes:     51 * 1024 * 1024,
		CoursePhaseID: &coursePhaseID,
	})
	assert.ErrorIs(suite.T(), err, ErrInvalidInput)
}

// TestCreateFileFromStorageKey_OversizedUploadIsDeleted verifies an upload rejected at
// completion does not stay in storage. Regression test for GHSA-c898-5f75-f8qp.
func (suite *StorageServiceTestSuite) TestCreateFileFromStorageKey_OversizedUploadIsDeleted() {
	coursePhaseID := uuid.New()
	presigned, err := suite.service.PresignUpload(suite.ctx, PresignUploadRequest{
		Filename:      "big.pdf",
		ContentType:   "application/pdf",
		CoursePhaseID: &coursePhaseID,
	})
	require.NoError(suite.T(), err)

	mock := suite.mockAdapter.(*storage.MockStorageAdapter)
	mock.GetMetadataFunc = func(_ context.Context, storageKey string) (*storage.FileMetadata, error) {
		return &storage.FileMetadata{StorageKey: storageKey, Size: 60 * 1024 * 1024, ContentType: "application/pdf"}, nil
	}
	var deletedKeys []string
	mock.DeleteFunc = func(_ context.Context, storageKey string) error {
		deletedKeys = append(deletedKeys, storageKey)
		return nil
	}
	defer func() {
		mock.GetMetadataFunc = nil
		mock.DeleteFunc = nil
	}()

	_, err = suite.service.CreateFileFromStorageKey(suite.ctx, CreateFileFromStorageKeyRequest{
		StorageKey:    presigned.StorageKey,
		CoursePhaseID: &coursePhaseID,
		ContentType:   "application/pdf",
	}, suite.testUserID, "")
	assert.ErrorIs(suite.T(), err, ErrInvalidInput)
	assert.Equal(suite.T(), []string{presigned.StorageKey}, deletedKeys)
	assert.False(suite.T(), suite.pendingUploadExists(presigned.StorageKey))
}

func (suite *StorageServiceTestSuite) TestCreateFileFromStorageKey_ClearsPendingUpload() {
	coursePhaseID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	presigned, err := suite.service.PresignUpload(suite.ctx, PresignUploadRequest{
		Filename:      "ok.pdf",
		ContentType:   "application/pdf",
		CoursePhaseID: &coursePhaseID,
	})
	require.NoError(suite.T(), err)

	_, err = suite.service.CreateFileFromStorageKey(suite.ctx, CreateFileFromStorageKeyRequest{
		StorageKey:    presigned.StorageKey,
		CoursePhaseID: &coursePhaseID,
		ContentType:   "application/pdf",
	}, suite.testUserID, "")
	require.NoError(suite.T(), err)
	assert.False(suite.T(), suite.pendingUploadExists(presigned.StorageKey))
}

// TestReclaimExpiredUploads verifies the reaper deletes expired unregistered uploads, keeps
// the objects of registered files and leaves fresh uploads alone.
func (suite *StorageServiceTestSuite) TestReclaimExpiredUploads() {
	coursePhaseID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	presign := func(name string) string {
		presigned, err := suite.service.PresignUpload(suite.ctx, PresignUploadRequest{
			Filename:      name,
			ContentType:   "application/pdf",
			CoursePhaseID: &coursePhaseID,
		})
		require.NoError(suite.T(), err)
		return presigned.StorageKey
	}
	abandonedKey := presign("abandoned.pdf")
	registeredKey := presign("registered.pdf")
	freshKey := presign("fresh.pdf")

	_, err := suite.service.CreateFileFromStorageKey(suite.ctx, CreateFileFromStorageKeyRequest{
		StorageKey:    registeredKey,
		CoursePhaseID: &coursePhaseID,
		ContentType:   "application/pdf",
	}, suite.testUserID, "")
	require.NoError(suite.T(), err)
	// Simulate an entry the completion could not clear.
	_, err = suite.service.conn.Exec(suite.ctx, "INSERT INTO pending_upload (storage_key) VALUES ($1)", registeredKey)
	require.NoError(suite.T(), err)
	_, err = suite.service.conn.Exec(suite.ctx, "UPDATE pending_upload SET created_at = NOW() - INTERVAL '1 day' WHERE storage_key = ANY($1)", []string{abandonedKey, registeredKey})
	require.NoError(suite.T(), err)

	mock := suite.mockAdapter.(*storage.MockStorageAdapter)
	var deletedKeys []string
	mock.DeleteFunc = func(_ context.Context, storageKey string) error {
		deletedKeys = append(deletedKeys, storageKey)
		return nil
	}
	defer func() { mock.DeleteFunc = nil }()

	reclaimed, err := suite.service.ReclaimExpiredUploads(suite.ctx, 100)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, reclaimed)
	assert.Equal(suite.T(), []string{abandonedKey}, deletedKeys)
	assert.False(suite.T(), suite.pendingUploadExists(abandonedKey))
	assert.False(suite.T(), suite.pendingUploadExists(registeredKey))
	assert.True(suite.T(), suite.pendingUploadExists(freshKey))
}

func TestStorageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(StorageServiceTestSuite))
}
