package profilePicture

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/jpeg"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/profilePicture/profilePictureDTO"
	"github.com/prompt-edu/prompt/servers/core/storage/files"
	log "github.com/sirupsen/logrus"
)

const (
	// Pictures are served from here; only the server writes to it.
	storageKeyPrefix = "profile-picture"
	// Presigned uploads land here and are never served.
	uploadStorageKeyPrefix = "profile-picture-upload"
	pictureFilename        = "profile-picture.jpg"
	pictureContentType     = "image/jpeg"
	pictureDescription     = "Profile picture"
	// The client crops and re-encodes to a square JPEG of this size.
	pictureSizePx = 512
	// A valid 512 px JPEG is far below the limit.
	maxPictureSizeBytes = 1024 * 1024
	// Clients cache the URLs slightly shorter than this, so a table re-renders without re-fetching.
	urlTTLSeconds = 3600
)

// ErrInvalidInput marks a request the client has to fix (HTTP 400).
// ErrNotFound marks a missing profile picture or upload (HTTP 404).
var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
)

// FileStore is the slice of the file storage service the profile pictures are kept in.
type FileStore interface {
	PresignUpload(ctx context.Context, req files.PresignUploadRequest) (*files.PresignUploadResponse, error)
	DownloadUpload(ctx context.Context, storageKey string) (io.ReadCloser, error)
	DeleteUpload(ctx context.Context, storageKey string) error
	StoreFile(ctx context.Context, req files.StoreFileRequest, content io.Reader, uploaderUserID, uploaderEmail string) (*files.FileResponse, error)
	DeleteFile(ctx context.Context, fileID uuid.UUID, hardDelete bool) error
	GetDownloadURL(ctx context.Context, storageKey string, ttlSeconds int) (string, error)
}

// Uploader identifies the logged-in user a picture belongs to.
type Uploader struct {
	UserID          uuid.UUID
	UniversityLogin string
	Email           string
}

type ProfilePictureService struct {
	queries db.Queries
	conn    *pgxpool.Pool
	files   FileStore
}

func NewProfilePictureService(queries db.Queries, conn *pgxpool.Pool, fileStore FileStore) *ProfilePictureService {
	return &ProfilePictureService{
		queries: queries,
		conn:    conn,
		files:   fileStore,
	}
}

// PresignUpload returns an upload target scoped to the caller, valid for the configured upload TTL.
func (s *ProfilePictureService) PresignUpload(ctx context.Context, userID uuid.UUID) (profilePictureDTO.PresignedUpload, error) {
	presigned, err := s.files.PresignUpload(ctx, files.PresignUploadRequest{
		Filename:         pictureFilename,
		ContentType:      pictureContentType,
		StorageKeyPrefix: ownerKeyPrefix(uploadStorageKeyPrefix, userID),
		Description:      pictureDescription,
	})
	if err != nil {
		return profilePictureDTO.PresignedUpload{}, fmt.Errorf("failed to presign profile picture upload: %w", err)
	}
	return profilePictureDTO.PresignedUpload{UploadURL: presigned.UploadURL, StorageKey: presigned.StorageKey}, nil
}

// CompleteUpload checks an uploaded picture and makes it the uploader's profile picture, replacing
// and deleting the previous one. The presigned URL stays writable until it expires, so the checked
// bytes are stored under a key that was never handed out, and the upload itself is removed.
func (s *ProfilePictureService) CompleteUpload(ctx context.Context, uploader Uploader, uploadKey string) (profilePictureDTO.ProfilePicture, error) {
	if err := validateUploadKeyOwner(uploadKey, uploader.UserID); err != nil {
		return profilePictureDTO.ProfilePicture{}, err
	}

	content, err := s.readUploadedPicture(ctx, uploadKey)
	if err != nil {
		// A storage hiccup says nothing about the picture, so that upload can be completed again
		if errors.Is(err, ErrInvalidInput) {
			s.discardUpload(ctx, uploadKey)
		}
		return profilePictureDTO.ProfilePicture{}, err
	}

	file, err := s.files.StoreFile(ctx, files.StoreFileRequest{
		StorageKeyPrefix: ownerKeyPrefix(storageKeyPrefix, uploader.UserID),
		Filename:         pictureFilename,
		ContentType:      pictureContentType,
		Description:      pictureDescription,
	}, bytes.NewReader(content), uploader.UserID.String(), uploader.Email)
	if err != nil {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("failed to store profile picture: %w", err)
	}

	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	// Concurrent uploads of one user are serialized, so each one sees the file it replaces and
	// no replaced file is left behind unreferenced.
	tx, err := s.conn.Begin(ctxWithTimeout)
	if err != nil {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctxWithTimeout)
	txQueries := s.queries.WithTx(tx)

	if err := txQueries.LockProfilePictureOfUser(ctxWithTimeout, uploader.UserID); err != nil {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("failed to lock profile picture: %w", err)
	}

	previous, err := txQueries.GetProfilePictureByUserID(ctxWithTimeout, uploader.UserID)
	hasPrevious := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("failed to load current profile picture: %w", err)
	}

	picture, err := txQueries.UpsertProfilePicture(ctxWithTimeout, db.UpsertProfilePictureParams{
		UserID:          uploader.UserID,
		UniversityLogin: pgtype.Text{String: uploader.UniversityLogin, Valid: uploader.UniversityLogin != ""},
		FileID:          file.ID,
	})
	if err != nil {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("failed to save profile picture: %w", err)
	}

	if err := tx.Commit(ctxWithTimeout); err != nil {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("failed to commit profile picture: %w", err)
	}

	s.discardUpload(ctx, uploadKey)
	if hasPrevious {
		s.discardFile(ctx, previous.FileID)
	}

	url, err := s.files.GetDownloadURL(ctx, file.StorageKey, urlTTLSeconds)
	if err != nil {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("failed to create profile picture URL: %w", err)
	}
	return profilePictureDTO.ProfilePicture{URL: url, UpdatedAt: picture.UpdatedAt.Time}, nil
}

// GetOwnPicture returns the caller's profile picture, or ErrNotFound if they have none.
func (s *ProfilePictureService) GetOwnPicture(ctx context.Context, userID uuid.UUID) (profilePictureDTO.ProfilePicture, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	picture, err := s.queries.GetProfilePictureByUserID(ctxWithTimeout, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("no profile picture: %w", ErrNotFound)
	}
	if err != nil {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("failed to load profile picture: %w", err)
	}

	url, err := s.files.GetDownloadURL(ctx, picture.StorageKey, urlTTLSeconds)
	if err != nil {
		return profilePictureDTO.ProfilePicture{}, fmt.Errorf("failed to create profile picture URL: %w", err)
	}
	return profilePictureDTO.ProfilePicture{URL: url, UpdatedAt: picture.UpdatedAt.Time}, nil
}

// DeleteOwnPicture removes the caller's profile picture and its stored file.
func (s *ProfilePictureService) DeleteOwnPicture(ctx context.Context, userID uuid.UUID) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	tx, err := s.conn.Begin(ctxWithTimeout)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctxWithTimeout)
	txQueries := s.queries.WithTx(tx)

	if err := txQueries.LockProfilePictureOfUser(ctxWithTimeout, userID); err != nil {
		return fmt.Errorf("failed to lock profile picture: %w", err)
	}

	picture, err := txQueries.GetProfilePictureByUserID(ctxWithTimeout, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("no profile picture: %w", ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("failed to load profile picture: %w", err)
	}

	if err := txQueries.DeleteProfilePictureByUserID(ctxWithTimeout, userID); err != nil {
		return fmt.Errorf("failed to delete profile picture: %w", err)
	}

	if err := tx.Commit(ctxWithTimeout); err != nil {
		return fmt.Errorf("failed to commit profile picture removal: %w", err)
	}

	s.discardFile(ctx, picture.FileID)
	return nil
}

// LookupPictureURLs resolves presigned URLs for every requested id that has a profile picture.
func (s *ProfilePictureService) LookupPictureURLs(ctx context.Context, req profilePictureDTO.LookupRequest) (profilePictureDTO.ProfilePictureURLs, error) {
	if err := validateLookupRequest(req); err != nil {
		return profilePictureDTO.ProfilePictureURLs{}, err
	}

	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	result := profilePictureDTO.ProfilePictureURLs{
		Users:                map[uuid.UUID]string{},
		Students:             map[uuid.UUID]string{},
		CourseParticipations: map[uuid.UUID]string{},
	}
	// The same picture often appears under several ids, so each key is presigned once.
	urlsByStorageKey := map[string]string{}
	presign := func(storageKey string) (string, error) {
		if url, ok := urlsByStorageKey[storageKey]; ok {
			return url, nil
		}
		url, err := s.files.GetDownloadURL(ctx, storageKey, urlTTLSeconds)
		if err != nil {
			return "", fmt.Errorf("failed to create profile picture URL: %w", err)
		}
		urlsByStorageKey[storageKey] = url
		return url, nil
	}

	if len(req.UserIDs) > 0 {
		rows, err := s.queries.GetProfilePictureStorageKeysByUserIDs(ctxWithTimeout, req.UserIDs)
		if err != nil {
			return profilePictureDTO.ProfilePictureURLs{}, fmt.Errorf("failed to look up user profile pictures: %w", err)
		}
		for _, row := range rows {
			if result.Users[row.UserID], err = presign(row.StorageKey); err != nil {
				return profilePictureDTO.ProfilePictureURLs{}, err
			}
		}
	}

	if len(req.StudentIDs) > 0 {
		rows, err := s.queries.GetProfilePictureStorageKeysByStudentIDs(ctxWithTimeout, req.StudentIDs)
		if err != nil {
			return profilePictureDTO.ProfilePictureURLs{}, fmt.Errorf("failed to look up student profile pictures: %w", err)
		}
		for _, row := range rows {
			if result.Students[row.StudentID], err = presign(row.StorageKey); err != nil {
				return profilePictureDTO.ProfilePictureURLs{}, err
			}
		}
	}

	if len(req.CourseParticipationIDs) > 0 {
		rows, err := s.queries.GetProfilePictureStorageKeysByCourseParticipationIDs(ctxWithTimeout, req.CourseParticipationIDs)
		if err != nil {
			return profilePictureDTO.ProfilePictureURLs{}, fmt.Errorf("failed to look up course participation profile pictures: %w", err)
		}
		for _, row := range rows {
			if result.CourseParticipations[row.CourseParticipationID], err = presign(row.StorageKey); err != nil {
				return profilePictureDTO.ProfilePictureURLs{}, err
			}
		}
	}

	return result, nil
}

// readUploadedPicture reads an upload once and checks what the client promised: a square JPEG of
// pictureSizePx within the size limit. The browser-side crop produces exactly that, so anything
// else was not uploaded through PROMPT.
func (s *ProfilePictureService) readUploadedPicture(ctx context.Context, uploadKey string) ([]byte, error) {
	reader, err := s.files.DownloadUpload(ctx, uploadKey)
	if errors.Is(err, files.ErrNotFound) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, err.Error())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded profile picture: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("failed to close profile picture reader")
		}
	}()

	content, err := io.ReadAll(io.LimitReader(reader, maxPictureSizeBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded profile picture: %w", err)
	}
	if len(content) > maxPictureSizeBytes {
		return nil, fmt.Errorf("profile picture exceeds %d bytes: %w", maxPictureSizeBytes, ErrInvalidInput)
	}

	config, err := jpeg.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("profile picture is not a JPEG image: %w", ErrInvalidInput)
	}
	if config.Width != pictureSizePx || config.Height != pictureSizePx {
		return nil, fmt.Errorf("profile picture is %dx%d px, expected %dx%d: %w", config.Width, config.Height, pictureSizePx, pictureSizePx, ErrInvalidInput)
	}
	return content, nil
}

// discardUpload removes an upload that was rejected or stored again. Failures only leave an
// unreferenced object behind, which is never served.
func (s *ProfilePictureService) discardUpload(ctx context.Context, uploadKey string) {
	if err := s.files.DeleteUpload(ctx, uploadKey); err != nil {
		log.WithError(err).WithField("uploadKey", uploadKey).Warn("failed to delete profile picture upload")
	}
}

// discardFile hard-deletes a picture file. Failures only leave an orphaned blob behind, so they
// are logged instead of failing the request that replaced or removed the picture.
func (s *ProfilePictureService) discardFile(ctx context.Context, fileID uuid.UUID) {
	if err := s.files.DeleteFile(ctx, fileID, true); err != nil {
		log.WithError(err).WithField("fileID", fileID).Warn("failed to delete profile picture file")
	}
}
