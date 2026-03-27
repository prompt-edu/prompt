package service

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

// CreateExportRecord atomically checks that the user is allowed to create a new export
// (no valid export exists, not rate-limited) and inserts the record.
// Uses SELECT ... FOR UPDATE to prevent concurrent duplicate exports.
func CreateExportRecord(c *gin.Context, subject sdk.SubjectIdentifiers) (privacyDTO.PrivacyExport, error) {
	tx, err := PrivacyServiceSingleton.conn.Begin(c)
	if err != nil {
		return privacyDTO.PrivacyExport{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(c) }()

	txQueries := PrivacyServiceSingleton.queries.WithTx(tx)

	latestExport, err := txQueries.GetLatestExportForUserForUpdate(c, subject.UserID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return privacyDTO.PrivacyExport{}, fmt.Errorf("failed to check existing exports: %w", err)
	}

	if err == nil {
		dto := privacyDTO.GetPrivacyExportDTOFromDBModel(latestExport)
		if time.Now().Before(dto.ValidUntil) {
			return privacyDTO.PrivacyExport{}, fmt.Errorf("an export already exists and is valid until %s", dto.ValidUntil)
		}
		rateLimitEnd := dto.DateCreated.Add(exportRateLimit())
		if time.Now().Before(rateLimitEnd) {
			return privacyDTO.PrivacyExport{}, fmt.Errorf("rate limited until %s", rateLimitEnd)
		}
	}

	exp, err := txQueries.CreateNewExport(c, db.CreateNewExportParams{
		ID:         uuid.New(),
		UserID:     subject.UserID,
		StudentID:  pgtype.UUID{Bytes: subject.StudentID, Valid: subject.StudentID != uuid.Nil},
		Status:     db.ExportStatusPending,
		ValidUntil: pgtype.Timestamptz{Time: time.Now().Add(exportExpiry()), Valid: true},
	})
	if err != nil {
		return privacyDTO.PrivacyExport{}, fmt.Errorf("failed to create export record: %w", err)
	}

	if err := tx.Commit(c); err != nil {
		return privacyDTO.PrivacyExport{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return privacyDTO.GetPrivacyExportDTOFromDBModel(exp), nil
}

func CreateExportRecordDoc(c *gin.Context, exportID uuid.UUID, sourceName string) (db.PrivacyExportDocument, error) {
	return PrivacyServiceSingleton.queries.CreateNewExportDoc(c, db.CreateNewExportDocParams{
		ID:         uuid.New(),
		ExportID:   pgtype.UUID{Bytes: exportID, Valid: true},
		SourceName: sourceName,
		ObjectKey:  privacyexport.MakeObjectURL(exportID.String(), sourceName),
		Status:     db.ExportStatusPending,
	})
}

type ServiceExportRequest struct {
  APIURL    string
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

func PrepareExportRecordDoc(c *gin.Context, exportID uuid.UUID, sourceName string, apiURL string) (ServiceExportRequest, error) {

  dbDoc, errDoc := CreateExportRecordDoc(c, exportID, sourceName)
  if errDoc != nil {
  	return ServiceExportRequest{}, errDoc
  }

  url, errUrl := privacyexport.GetUploadURL(c, exportID.String(), sourceName)
  if errUrl != nil {
    if setErr := SetExportDocStatus(c, dbDoc.ID, db.ExportStatusFailed); setErr != nil {
      log.WithError(setErr).Error("failed to mark export doc as failed after presign error")
    }
    return ServiceExportRequest{}, errUrl
  }

  return ServiceExportRequest{
    PresignedUploadURL: url, 
    ExportDoc: privacyDTO.GetPrivacyExportDocDTOFromDBModel(dbDoc), 
    APIURL: apiURL,
    Result: Pending,
  }, nil
}


func GetExportWithDocs(c *gin.Context, exportID uuid.UUID) (privacyDTO.PrivacyExport, error) {
	dbExp, err := PrivacyServiceSingleton.queries.GetExportRecordByIDWithDocs(c, exportID)
	if err != nil {
		return privacyDTO.PrivacyExport{}, err
	}

	exp, err := privacyDTO.GetPrivacyExportWithDocsDTOFromDBModel(dbExp)
	if err != nil {
		return privacyDTO.PrivacyExport{}, err
	}
	return exp, nil
}

func GetLatestExportWithDocs(c *gin.Context, userID uuid.UUID) (*privacyDTO.PrivacyExport, error) {
	dbExp, err := PrivacyServiceSingleton.queries.GetLatestExportForUserWithDocs(c, userID)
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
	ExportReadyForNew   ExportAvailability = iota // no export or expired + not rate limited
	ExportExistsAndValid                          // export exists and files are still available, (implies rate limited)
	ExportRateLimited                             // export expired but still within rate limit window
)

func GetExportAvailability(c *gin.Context, userID uuid.UUID) (ExportAvailability, *privacyDTO.PrivacyExport, error) {
	exp, err := GetLatestExportWithDocs(c, userID)
	if err != nil {
		return ExportReadyForNew, nil, err
	}

	if exp == nil {
		return ExportReadyForNew, nil, nil
	}

	if time.Now().Before(exp.ValidUntil) {
		return ExportExistsAndValid, exp, nil
	}

	rateLimitEnd := exp.DateCreated.Add(exportRateLimit())
	if time.Now().Before(rateLimitEnd) {
		return ExportRateLimited, exp, nil
	}

	return ExportReadyForNew, nil, nil
}

func GetAllExports(c *gin.Context) ([]privacyDTO.AdminPrivacyExport, error) {
	dbExports, err := PrivacyServiceSingleton.queries.GetAllExports(c)
	if err != nil {
		return nil, err
	}

	exports := make([]privacyDTO.AdminPrivacyExport, 0, len(dbExports))
	for _, dbExp := range dbExports {
		exports = append(exports, privacyDTO.GetAdminPrivacyExportDTOFromDBModel(dbExp))
	}
	return exports, nil
}

func RateLimitEndForExport(exp privacyDTO.PrivacyExport) time.Time {
	return exp.DateCreated.Add(exportRateLimit())
}

func SetExportDocStatus(c *gin.Context, docID uuid.UUID, status db.ExportStatus) error {
  _, err := PrivacyServiceSingleton.queries.SetExportDocStatus(c, db.SetExportDocStatusParams{
    ID: docID,
    Status: status,
  })
  return err
}

func SetExportStatus(c *gin.Context, exportID uuid.UUID, status db.ExportStatus) error {
  _, err := PrivacyServiceSingleton.queries.SetExportStatus(c, db.SetExportStatusParams{
    ID: exportID,
    Status: status,
  })
  return err
}

func UpdateExportStatus(err error, c *gin.Context, exportID uuid.UUID) {
  if err != nil {
    if setErr := SetExportStatus(c, exportID, db.ExportStatusFailed); setErr != nil {
      log.WithError(setErr).Error("failed to set export status to failed")
    }
  } else {
    if setErr := SetExportStatus(c, exportID, db.ExportStatusComplete); setErr != nil {
      log.WithError(setErr).Error("failed to set export status to complete")
    }
  }
}

func UpdateExportDocStatus(err error, c *gin.Context, exportDocID uuid.UUID) {
  var targetStatus db.ExportStatus
  if err != nil {
    targetStatus = db.ExportStatusFailed
  } else {
    targetStatus = db.ExportStatusComplete
  }

  if _, setErr := PrivacyServiceSingleton.queries.SetExportDocStatus(c, db.SetExportDocStatusParams{
    ID: exportDocID, 
    Status: targetStatus,
  }) ; setErr != nil {
    log.WithError(setErr).Error("failed to update export doc result")
  }
}


func UpdateExportDocFileSize(c *gin.Context, exportDocID uuid.UUID) {
  objectKey, keyErr := PrivacyServiceSingleton.queries.GetExportDocObjectKey(c, exportDocID)
  if keyErr != nil {
    if setErr := SetExportDocStatus(c, exportDocID, db.ExportStatusComplete); setErr != nil {
      log.WithError(setErr).Error("failed to set export doc status to complete")
    }
    return
  }

  fileSize, sizeErr := privacyexport.GetFileSize(c, objectKey)
  if sizeErr != nil {
      log.WithError(sizeErr).Error("failed to get file size from s3 obj")
      return
  }

  if _, setErr := PrivacyServiceSingleton.queries.SetExportDocFileSize(c, db.SetExportDocFileSizeParams{
    ID: exportDocID, 
    FileSize: pgtype.Int8{ Int64: fileSize, Valid: true },
  }) ; setErr != nil {
    log.WithError(setErr).Error("failed to update export doc result")
  }
}
