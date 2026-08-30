package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	"github.com/prompt-edu/prompt/servers/core/storage/privacyexport"
	log "github.com/sirupsen/logrus"
)

// exportRateLimit : how long a user must wait before requesting a new export
func exportRateLimit() time.Duration {
	if val := os.Getenv("PRIVACY_EXPORT_RATE_LIMIT_DAYS"); val != "" {
		if days, err := strconv.Atoi(val); err == nil && days > 0 {
			return time.Duration(days) * 24 * time.Hour
		}
	}
	return 30 * 24 * time.Hour
}

// exportExpiry : how long export files remain available for download
func exportExpiry() time.Duration {
	if val := os.Getenv("PRIVACY_EXPORT_EXPIRY_DAYS"); val != "" {
		if days, err := strconv.Atoi(val); err == nil && days > 0 {
			return time.Duration(days) * 24 * time.Hour
		}
	}
	return 7 * 24 * time.Hour
}

// CreateExportRecord inserts a new export record in pending state for the given subject.
// Callers must run ValidateNoValidExportExists and ValidateNotRateLimited first.
func (s *PrivacyService) CreateExportRecord(c context.Context, subjectIdentifiers sdk.SubjectIdentifiers) (privacyDTO.PrivacyExport, error) {
	now := time.Now()
	exp, err := s.queries.CreateNewExport(c, db.CreateNewExportParams{
		ID:                   uuid.New(),
		UserID:               pgtype.UUID{Bytes: subjectIdentifiers.UserID, Valid: subjectIdentifiers.UserID != uuid.Nil},
		StudentID:            pgtype.UUID{Bytes: subjectIdentifiers.StudentID, Valid: subjectIdentifiers.StudentID != uuid.Nil},
		Status:               db.ExportStatusPending,
		ValidUntil:           pgtype.Timestamptz{Time: now.Add(exportExpiry()), Valid: true},
		NextRequestAllowedAt: pgtype.Timestamptz{Time: now.Add(exportRateLimit()), Valid: true},
	})
	if err != nil {
		return privacyDTO.PrivacyExport{}, fmt.Errorf("failed to create export record: %w", err)
	}
	return privacyDTO.GetPrivacyExportDTOFromDBModel(exp), nil
}

func (s *PrivacyService) CreateExportRecordDoc(c context.Context, exportID uuid.UUID, sourceName string) (db.PrivacyExportDocument, error) {
	return s.queries.CreateNewExportDoc(c, db.CreateNewExportDocParams{
		ID:         uuid.New(),
		ExportID:   pgtype.UUID{Bytes: exportID, Valid: true},
		SourceName: sourceName,
		ObjectKey:  privacyexport.MakeObjectURL(exportID.String(), sourceName),
		Status:     db.ExportStatusPending,
	})
}

type ServiceExportRequest struct {
	APIURL             string
	PresignedUploadURL string
	ExportDoc          privacyDTO.PrivacyExportDocument
	Result             ExportResult
}

type ExportResult int

const (
	Successful ExportResult = iota
	SuccessfulNoData
	Failed
	Pending
)

func (s *PrivacyService) PrepareExportRecordDoc(c context.Context, exportID uuid.UUID, sourceName string, apiURL string) (ServiceExportRequest, error) {

	dbDoc, errDoc := s.CreateExportRecordDoc(c, exportID, sourceName)
	if errDoc != nil {
		return ServiceExportRequest{}, errDoc
	}

	url, errUrl := privacyexport.GetUploadURL(c, exportID.String(), sourceName)
	if errUrl != nil {
		if setErr := s.SetExportDocStatus(c, dbDoc.ID, db.ExportStatusFailed); setErr != nil {
			log.WithError(setErr).Error("failed to mark export doc as failed after presign error")
		}
		return ServiceExportRequest{}, errUrl
	}

	return ServiceExportRequest{
		PresignedUploadURL: url,
		ExportDoc:          privacyDTO.GetPrivacyExportDocDTOFromDBModel(dbDoc),
		APIURL:             apiURL,
		Result:             Pending,
	}, nil
}

func (s *PrivacyService) GetExportWithDocs(c context.Context, exportID uuid.UUID) (privacyDTO.PrivacyExport, error) {
	dbExp, err := s.queries.GetExportRecordByIDWithDocs(c, exportID)
	if err != nil {
		return privacyDTO.PrivacyExport{}, err
	}

	exp, err := privacyDTO.GetPrivacyExportWithDocsDTOFromDBModel(dbExp)
	if err != nil {
		return privacyDTO.PrivacyExport{}, err
	}
	return exp, nil
}

func (s *PrivacyService) GetLatestExportWithDocs(c context.Context, userID uuid.UUID) (*privacyDTO.PrivacyExport, error) {
	dbExp, err := s.queries.GetLatestExportForUserWithDocs(c, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	exp, err := privacyDTO.GetPrivacyExportWithDocsDTOFromDBModel(dbExp)
	if err != nil {
		return nil, err
	}
	return &exp, nil
}

type ExportAvailability int

const (
	ExportReadyForNew    ExportAvailability = iota // no export or expired + not rate limited
	ExportExistsAndValid                           // export exists and files are still available, (implies rate limited)
	ExportRateLimited                              // export expired but still within rate limit window
)

func (s *PrivacyService) GetExportAvailability(c context.Context, userID uuid.UUID) (ExportAvailability, *privacyDTO.PrivacyExport, error) {
	exp, err := s.GetLatestExportWithDocs(c, userID)
	if err != nil {
		return ExportReadyForNew, nil, err
	}

	if exp == nil {
		return ExportReadyForNew, nil, nil
	}

	if time.Now().Before(exp.ValidUntil) {
		return ExportExistsAndValid, exp, nil
	}

	if time.Now().Before(exp.NextRequestAllowedAt) {
		return ExportRateLimited, exp, nil
	}

	return ExportReadyForNew, nil, nil
}

func (s *PrivacyService) GetAllExports(c context.Context) ([]privacyDTO.AdminPrivacyExport, error) {
	dbExports, err := s.queries.GetAllExports(c)
	if err != nil {
		return nil, err
	}

	exports := make([]privacyDTO.AdminPrivacyExport, 0, len(dbExports))
	for _, dbExp := range dbExports {
		dto, err := privacyDTO.GetAdminPrivacyExportDTOFromDBModel(dbExp)
		if err != nil {
			return nil, err
		}
		exports = append(exports, dto)
	}
	return exports, nil
}

func RateLimitEndForExport(exp privacyDTO.PrivacyExport) time.Time {
	return exp.NextRequestAllowedAt
}

func (s *PrivacyService) ResetExportRateLimit(ctx context.Context, exportID uuid.UUID) error {
	return s.queries.ResetExportNextRequestAllowedAt(ctx, exportID)
}

func (s *PrivacyService) SetExportDocStatus(c context.Context, docID uuid.UUID, status db.ExportStatus) error {
	_, err := s.queries.SetExportDocStatus(c, db.SetExportDocStatusParams{
		ID:     docID,
		Status: status,
	})
	return err
}

func exportResultToDBStatus(result ExportResult) db.ExportStatus {
	switch result {
	case Successful:
		return db.ExportStatusComplete
	case SuccessfulNoData:
		return db.ExportStatusNoData
	default:
		return db.ExportStatusFailed
	}
}

func (s *PrivacyService) SetExportStatus(c context.Context, exportID uuid.UUID, status db.ExportStatus) error {
	_, err := s.queries.SetExportStatus(c, db.SetExportStatusParams{
		ID:     exportID,
		Status: status,
	})
	return err
}

func (s *PrivacyService) UpdateExportStatus(err error, c context.Context, exportID uuid.UUID) {
	if err != nil {
		if setErr := s.SetExportStatus(c, exportID, db.ExportStatusFailed); setErr != nil {
			log.WithError(setErr).Error("failed to set export status to failed")
		}
	} else {
		if setErr := s.SetExportStatus(c, exportID, db.ExportStatusComplete); setErr != nil {
			log.WithError(setErr).Error("failed to set export status to complete")
		}
	}
}

func (s *PrivacyService) UpdateExportDocStatus(err error, c context.Context, exportDocID uuid.UUID) {
	var targetStatus db.ExportStatus
	if err != nil {
		targetStatus = db.ExportStatusFailed
	} else {
		targetStatus = db.ExportStatusComplete
	}

	if _, setErr := s.queries.SetExportDocStatus(c, db.SetExportDocStatusParams{
		ID:     exportDocID,
		Status: targetStatus,
	}); setErr != nil {
		log.WithError(setErr).Error("failed to update export doc result")
	}
}

func (s *PrivacyService) UpdateExportDocFileSize(c context.Context, exportDocID uuid.UUID) {
	objectKey, keyErr := s.queries.GetExportDocObjectKey(c, exportDocID)
	if keyErr != nil {
		if setErr := s.SetExportDocStatus(c, exportDocID, db.ExportStatusComplete); setErr != nil {
			log.WithError(setErr).Error("failed to set export doc status to complete")
		}
		return
	}

	fileSize, sizeErr := privacyexport.GetFileSize(c, objectKey)
	if sizeErr != nil {
		log.WithError(sizeErr).Error("failed to get file size from s3 obj")
		return
	}

	if _, setErr := s.queries.SetExportDocFileSize(c, db.SetExportDocFileSizeParams{
		ID:       exportDocID,
		FileSize: pgtype.Int8{Int64: fileSize, Valid: true},
	}); setErr != nil {
		log.WithError(setErr).Error("failed to update export doc result")
	}
}
