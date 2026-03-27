package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/storage"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	log "github.com/sirupsen/logrus"
)

// StorageService provides business logic for file operations
type StorageService struct {
	queries        db.Queries
	conn           *pgxpool.Pool
	storageAdapter storage.StorageAdapter
	maxFileSize    int64
	allowedTypes   []string
}

var StorageServiceSingleton *StorageService

// FileUploadRequest represents a file upload request
type FileUploadRequest struct {
	File           *multipart.FileHeader
	UploaderUserID string
	UploaderEmail  string
	CoursePhaseID  *uuid.UUID
	Description    string
	Tags           []string
}

type PresignUploadRequest struct {
	Filename      string
	ContentType   string
	CoursePhaseID *uuid.UUID
	Description   string
	Tags          []string
}

type PresignUploadResponse struct {
	UploadURL  string `json:"uploadUrl"`
	StorageKey string `json:"storageKey"`
}

type CreateFileFromStorageKeyRequest struct {
	StorageKey       string
	OriginalFilename string
	ContentType      string
	CoursePhaseID    *uuid.UUID
	Description      string
	Tags             []string
}

// FileResponse represents a file in API responses
type FileResponse struct {
	ID               uuid.UUID  `json:"id"`
	Filename         string     `json:"filename"`
	OriginalFilename string     `json:"originalFilename"`
	ContentType      string     `json:"contentType"`
	SizeBytes        int64      `json:"sizeBytes"`
	StorageKey       string     `json:"storageKey"`
	DownloadURL      string     `json:"downloadUrl"`
	UploadedByUserID string     `json:"uploadedByUserId"`
	UploadedByEmail  string     `json:"uploadedByEmail,omitempty"`
	CoursePhaseID    *uuid.UUID `json:"coursePhaseId,omitempty"`
	Description      string     `json:"description,omitempty"`
	Tags             []string   `json:"tags,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// NewStorageService creates a new storage service instance
func NewStorageService(queries db.Queries, conn *pgxpool.Pool, adapter storage.StorageAdapter, maxFileSizeMB int64, allowedTypes []string) *StorageService {
	return &StorageService{
		queries:        queries,
		conn:           conn,
		storageAdapter: adapter,
		maxFileSize:    maxFileSizeMB * 1024 * 1024, // Convert MB to bytes
		allowedTypes:   allowedTypes,
	}
}

// UploadFile handles the complete file upload process
func (s *StorageService) UploadFile(ctx context.Context, req FileUploadRequest) (*FileResponse, error) {
	// Validate file size
	if req.File.Size > s.maxFileSize {
		return nil, fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes", req.File.Size, s.maxFileSize)
	}

	// Validate file type if restrictions are set
	if len(s.allowedTypes) > 0 && !s.isAllowedType(req.File.Header.Get("Content-Type")) {
		return nil, fmt.Errorf("file type %s is not allowed", req.File.Header.Get("Content-Type"))
	}

	// Open the uploaded file
	file, err := req.File.Open()
	if err != nil {
		log.WithError(err).Error("Failed to open uploaded file")
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("Failed to close uploaded file")
		}
	}()

	// Generate a unique filename to prevent collisions
	ext := filepath.Ext(req.File.Filename)
	safeOriginal := sanitizeFilename(req.File.Filename)
	if safeOriginal == "" {
		safeOriginal = "file" + ext
	}
	uniqueFilename := fmt.Sprintf("%s-%s", uuid.New().String(), safeOriginal)
	storageKey := buildStorageKey(req.CoursePhaseID, uniqueFilename)

	// Upload to storage backend
	uploadResult, err := s.storageAdapter.Upload(ctx, storageKey, req.File.Header.Get("Content-Type"), file)
	if err != nil {
		log.WithError(err).Error("Failed to upload file to storage backend")
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Get storage provider from environment
	storageProvider := sdkUtils.GetEnv("STORAGE_PROVIDER", "seaweedfs")

	// Save file metadata to database
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	var coursePhaseIDPgtype pgtype.UUID
	if req.CoursePhaseID != nil {
		coursePhaseIDPgtype = pgtype.UUID{Bytes: *req.CoursePhaseID, Valid: true}
	}

	var emailPgtype pgtype.Text
	if req.UploaderEmail != "" {
		emailPgtype = pgtype.Text{String: req.UploaderEmail, Valid: true}
	}

	var descriptionPgtype pgtype.Text
	if req.Description != "" {
		descriptionPgtype = pgtype.Text{String: req.Description, Valid: true}
	}

	fileRecord, err := s.queries.CreateFile(ctxWithTimeout, db.CreateFileParams{
		Filename:         uniqueFilename,
		OriginalFilename: req.File.Filename,
		ContentType:      req.File.Header.Get("Content-Type"),
		SizeBytes:        uploadResult.Size,
		StorageKey:       uploadResult.StorageKey,
		StorageProvider:  storageProvider,
		UploadedByUserID: req.UploaderUserID,
		UploadedByEmail:  emailPgtype,
		CoursePhaseID:    coursePhaseIDPgtype,
		Description:      descriptionPgtype,
		Tags:             req.Tags,
	})

	if err != nil {
		log.WithError(err).Error("Failed to save file metadata to database")
		// Attempt to delete the uploaded file from storage
		_ = s.storageAdapter.Delete(ctx, uploadResult.StorageKey)
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	log.WithFields(log.Fields{
		"fileId":     fileRecord.ID,
		"filename":   uniqueFilename,
		"size":       uploadResult.Size,
		"uploadedBy": req.UploaderUserID,
	}).Info("File uploaded successfully")

	return s.convertToFileResponse(ctx, fileRecord), nil
}

// PresignUpload creates a presigned upload URL for a file
func (s *StorageService) PresignUpload(ctx context.Context, req PresignUploadRequest) (*PresignUploadResponse, error) {
	if req.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	if req.ContentType == "" {
		return nil, fmt.Errorf("content type is required")
	}

	if len(s.allowedTypes) > 0 && !s.isAllowedType(req.ContentType) {
		return nil, fmt.Errorf("file type %s is not allowed", req.ContentType)
	}

	safeOriginal := sanitizeFilename(req.Filename)
	if safeOriginal == "" {
		ext := filepath.Ext(req.Filename)
		safeOriginal = "file" + ext
	}
	uniqueFilename := fmt.Sprintf("%s-%s", uuid.New().String(), safeOriginal)
	storageKey := buildStorageKey(req.CoursePhaseID, uniqueFilename)

	uploadURL, err := s.storageAdapter.GetUploadURL(ctx, storageKey, req.ContentType, presignUploadTTLSeconds())
	if err != nil {
		return nil, err
	}

	return &PresignUploadResponse{
		UploadURL:  uploadURL,
		StorageKey: storageKey,
	}, nil
}

// CreateFileFromStorageKey stores file metadata after a presigned upload completes.
func (s *StorageService) CreateFileFromStorageKey(ctx context.Context, req CreateFileFromStorageKeyRequest, uploaderUserID, uploaderEmail string) (*FileResponse, error) {
	if req.StorageKey == "" {
		return nil, fmt.Errorf("storage key is required")
	}

	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	existing, err := s.queries.GetFileByStorageKey(ctxWithTimeout, req.StorageKey)
	if err == nil {
		if existing.UploadedByUserID != uploaderUserID {
			return nil, fmt.Errorf("storage key already used by another user")
		}
		if req.CoursePhaseID != nil {
			if !existing.CoursePhaseID.Valid || existing.CoursePhaseID.Bytes != *req.CoursePhaseID {
				return nil, fmt.Errorf("storage key already used for a different course phase")
			}
		}
		return s.convertToFileResponse(ctx, existing), nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing file: %w", err)
	}

	if req.CoursePhaseID == nil {
		if strings.HasPrefix(req.StorageKey, "course-phase/") {
			return nil, fmt.Errorf("course phase id is required for course-phase storage keys")
		}
	} else {
		expectedPrefix := fmt.Sprintf("course-phase/%s/", req.CoursePhaseID.String())
		if !strings.HasPrefix(req.StorageKey, expectedPrefix) {
			return nil, fmt.Errorf("storage key does not match course phase")
		}
	}

	metadata, err := s.storageAdapter.GetMetadata(ctx, req.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata for storage key: %w", err)
	}

	contentType := metadata.ContentType
	if contentType == "" {
		contentType = req.ContentType
	}

	if contentType == "" {
		return nil, fmt.Errorf("content type is required")
	}

	if metadata.Size > s.maxFileSize {
		return nil, fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes", metadata.Size, s.maxFileSize)
	}

	if len(s.allowedTypes) > 0 && !s.isAllowedType(contentType) {
		return nil, fmt.Errorf("file type %s is not allowed", contentType)
	}

	uniqueFilename := filepath.Base(req.StorageKey)
	originalFilename := req.OriginalFilename
	if originalFilename == "" {
		originalFilename = uniqueFilename
	}

	storageProvider := sdkUtils.GetEnv("STORAGE_PROVIDER", "seaweedfs")

	var coursePhaseIDPgtype pgtype.UUID
	if req.CoursePhaseID != nil {
		coursePhaseIDPgtype = pgtype.UUID{Bytes: *req.CoursePhaseID, Valid: true}
	}

	var emailPgtype pgtype.Text
	if uploaderEmail != "" {
		emailPgtype = pgtype.Text{String: uploaderEmail, Valid: true}
	}

	var descriptionPgtype pgtype.Text
	if req.Description != "" {
		descriptionPgtype = pgtype.Text{String: req.Description, Valid: true}
	}

	fileRecord, err := s.queries.CreateFile(ctxWithTimeout, db.CreateFileParams{
		Filename:         uniqueFilename,
		OriginalFilename: originalFilename,
		ContentType:      contentType,
		SizeBytes:        metadata.Size,
		StorageKey:       req.StorageKey,
		StorageProvider:  storageProvider,
		UploadedByUserID: uploaderUserID,
		UploadedByEmail:  emailPgtype,
		CoursePhaseID:    coursePhaseIDPgtype,
		Description:      descriptionPgtype,
		Tags:             req.Tags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return s.convertToFileResponse(ctx, fileRecord), nil
}

// GetFileByID retrieves a file by its ID
func (s *StorageService) GetFileByID(ctx context.Context, fileID uuid.UUID) (*FileResponse, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	fileRecord, err := s.queries.GetFileByID(ctxWithTimeout, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	return s.convertToFileResponse(ctx, fileRecord), nil
}

// DownloadFile retrieves a file's content from storage
func (s *StorageService) DownloadFile(ctx context.Context, fileID uuid.UUID) (io.ReadCloser, string, error) {
	fileResp, err := s.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, "", err
	}

	reader, err := s.storageAdapter.Download(ctx, fileResp.StorageKey)
	if err != nil {
		log.WithError(err).WithField("fileId", fileID).Error("Failed to download file from storage")
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}

	return reader, fileResp.OriginalFilename, nil
}

// DeleteFile soft deletes a file and optionally removes it from storage
func (s *StorageService) DeleteFile(ctx context.Context, fileID uuid.UUID, hardDelete bool) error {
	fileResp, err := s.GetFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	if hardDelete {
		// Hard delete from database
		if err := s.queries.HardDeleteFile(ctxWithTimeout, fileID); err != nil {
			return fmt.Errorf("failed to delete file record: %w", err)
		}

		// Delete from storage backend
		if err := s.storageAdapter.Delete(ctx, fileResp.StorageKey); err != nil {
			log.WithError(err).WithField("fileId", fileID).Warn("File record deleted but storage deletion failed - orphaned storage object")
			return nil
		}

		log.WithField("fileId", fileID).Info("File hard deleted successfully")
	} else {
		// Soft delete (mark as deleted)
		if err := s.queries.SoftDeleteFile(ctxWithTimeout, fileID); err != nil {
			return fmt.Errorf("failed to soft delete file: %w", err)
		}

		log.WithField("fileId", fileID).Info("File soft deleted successfully")
	}

	return nil
}

// GetFilesByUploader retrieves all files uploaded by a specific user
func (s *StorageService) GetFilesByUploader(ctx context.Context, uploaderUserID string, limit, offset int32) ([]FileResponse, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	files, err := s.queries.GetFilesByUploader(ctxWithTimeout, db.GetFilesByUploaderParams{
		UploadedByUserID: uploaderUserID,
		Limit:            limit,
		Offset:           offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve files: %w", err)
	}

	responses := make([]FileResponse, len(files))
	for i, file := range files {
		responses[i] = *s.convertToFileResponse(ctx, file)
	}

	return responses, nil
}

// GetFilesByCoursePhaseID retrieves all files associated with a course phase
func (s *StorageService) GetFilesByCoursePhaseID(ctx context.Context, coursePhaseID uuid.UUID) ([]FileResponse, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coursePhaseIDPgtype := pgtype.UUID{Bytes: coursePhaseID, Valid: true}
	files, err := s.queries.GetFilesByCoursePhaseID(ctxWithTimeout, coursePhaseIDPgtype)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve files: %w", err)
	}

	responses := make([]FileResponse, len(files))
	for i, file := range files {
		responses[i] = *s.convertToFileResponse(ctx, file)
	}

	return responses, nil
}

func buildStorageKey(coursePhaseID *uuid.UUID, filename string) string {
	if coursePhaseID == nil {
		return filename
	}

	return fmt.Sprintf("course-phase/%s/%s", coursePhaseID.String(), filename)
}

func sanitizeFilename(filename string) string {
	sanitized := filepath.Base(filename)
	sanitized = strings.TrimSpace(sanitized)
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "..", "_")

	if sanitized == "" || sanitized == "." {
		return ""
	}

	return sanitized
}

// isAllowedType checks if the file type is allowed
func (s *StorageService) isAllowedType(contentType string) bool {
	if len(s.allowedTypes) == 0 {
		return true // No restrictions
	}

	normalizedContentType := strings.TrimSpace(contentType)
	if normalizedContentType == "" {
		return false
	}
	if mediaType, _, err := mime.ParseMediaType(normalizedContentType); err == nil {
		normalizedContentType = mediaType
	} else if separatorIndex := strings.Index(normalizedContentType, ";"); separatorIndex >= 0 {
		normalizedContentType = strings.TrimSpace(normalizedContentType[:separatorIndex])
	}

	for _, allowed := range s.allowedTypes {
		normalizedAllowedType := strings.TrimSpace(allowed)
		if normalizedAllowedType == "" {
			continue
		}
		if mediaType, _, err := mime.ParseMediaType(normalizedAllowedType); err == nil {
			normalizedAllowedType = mediaType
		} else if separatorIndex := strings.Index(normalizedAllowedType, ";"); separatorIndex >= 0 {
			normalizedAllowedType = strings.TrimSpace(normalizedAllowedType[:separatorIndex])
		}

		if strings.EqualFold(normalizedAllowedType, normalizedContentType) {
			return true
		}
	}
	return false
}

// convertToFileResponse converts a database file record to an API response
func (s *StorageService) convertToFileResponse(ctx context.Context, file db.File) *FileResponse {
	downloadURL, err := s.storageAdapter.GetURL(ctx, file.StorageKey, presignDownloadTTLSeconds())
	if err != nil {
		log.WithError(err).WithField("storageKey", file.StorageKey).Warn("Failed to generate download URL")
		downloadURL = ""
	}

	response := &FileResponse{
		ID:               file.ID,
		Filename:         file.Filename,
		OriginalFilename: file.OriginalFilename,
		ContentType:      file.ContentType,
		SizeBytes:        file.SizeBytes,
		StorageKey:       file.StorageKey,
		DownloadURL:      downloadURL,
		UploadedByUserID: file.UploadedByUserID,
		CreatedAt:        file.CreatedAt.Time,
		UpdatedAt:        file.UpdatedAt.Time,
	}

	if file.UploadedByEmail.Valid {
		response.UploadedByEmail = file.UploadedByEmail.String
	}

	if file.CoursePhaseID.Valid {
		phaseID := file.CoursePhaseID.Bytes
		phaseUUID := uuid.UUID(phaseID)
		response.CoursePhaseID = &phaseUUID
	}

	if file.Description.Valid {
		response.Description = file.Description.String
	}

	if len(file.Tags) > 0 {
		response.Tags = file.Tags
	}

	return response
}

func presignUploadTTLSeconds() int {
	return resolvePresignTTLSeconds("S3_PRESIGN_UPLOAD_TTL_SECONDS", 60)
}

func presignDownloadTTLSeconds() int {
	return resolvePresignTTLSeconds("S3_PRESIGN_DOWNLOAD_TTL_SECONDS", 30)
}

func resolvePresignTTLSeconds(envKey string, defaultTTL int) int {
	primaryValue := strings.TrimSpace(sdkUtils.GetEnv(envKey, ""))
	if primaryValue != "" {
		ttl, err := strconv.Atoi(primaryValue)
		if err == nil && ttl > 0 {
			return ttl
		}
	}

	legacyValue := strings.TrimSpace(sdkUtils.GetEnv("S3_PRESIGN_TTL_SECONDS", ""))
	if legacyValue != "" {
		ttl, err := strconv.Atoi(legacyValue)
		if err == nil && ttl > 0 {
			return ttl
		}
	}

	return defaultTTL
}
