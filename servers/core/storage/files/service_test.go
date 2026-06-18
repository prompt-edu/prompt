package files

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/storage"
	"github.com/stretchr/testify/assert"
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
	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../../database_dumps/storage_test.sql", func(pool *pgxpool.Pool) *db.Queries {
		return db.New(pool)
	})
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

	StorageServiceSingleton = suite.service
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

func TestStorageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(StorageServiceTestSuite))
}
