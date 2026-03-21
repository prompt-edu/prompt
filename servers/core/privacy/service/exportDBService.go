package service

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	"github.com/prompt-edu/prompt/servers/core/storage/privacyexport"
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

func CreateExportRecord(c *gin.Context, subjectIdentifiers sdk.SubjectIdentifiers) (privacyDTO.PrivacyExport, error){

  exp, err := PrivacyServiceSingleton.queries.CreateNewExport(c, db.CreateNewExportParams{
    ID:        uuid.New(),
    Userid:    subjectIdentifiers.UserID,
    Studentid: pgtype.UUID{Bytes: subjectIdentifiers.StudentID, Valid: subjectIdentifiers.StudentID != uuid.Nil},
    Status:    db.ExportStatusPending,
  })

  if err != nil {
    return privacyDTO.PrivacyExport{}, err
  }

  return privacyDTO.GetPrivacyExportDTOFromDBModel(exp), nil
}

func CreateExportRecordDoc(c *gin.Context, exportID uuid.UUID, sourceName string) (uuid.UUID, error) {

	docID := uuid.New()
	_, err := PrivacyServiceSingleton.queries.CreateNewExportDoc(c, db.CreateNewExportDocParams{
		ID:         docID,
		Exportid:   pgtype.UUID{Bytes: exportID, Valid: true},
		SourceName: sourceName,
		ObjectKey:  privacyexport.MakeObjectURL(exportID.String(), sourceName),
		Status:     db.ExportStatusPending,
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	return docID, nil
}


func PrepareExportRecordDoc(c *gin.Context, exportID uuid.UUID, sourceName string) (string, uuid.UUID, error) {

  docID, errDoc := CreateExportRecordDoc(c, exportID, sourceName)
  if errDoc != nil {
  	return "", uuid.UUID{}, errDoc
  }

  url, errUrl := privacyexport.GetUploadURL(c, exportID.String(), sourceName)
  if errUrl != nil {
    return "", uuid.UUID{}, errUrl
  }

  return url, docID, nil
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
	exp.AvailableUntil = exp.DateCreated.Add(exportExpiry())
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

	if time.Since(dbExp.DateCreated.Time) > exportRateLimit() {
		return nil, nil
	}

	exp, err := privacyDTO.GetPrivacyExportWithDocsDTOFromDBModel(dbExp)
	if err != nil {
		return nil, err
	}
	exp.AvailableUntil = exp.DateCreated.Add(exportExpiry())
	return &exp, nil
}

func SetExportDocStatus(c *gin.Context, docID uuid.UUID, status db.ExportStatus) {
  // TODO: what to do if error happens here
  PrivacyServiceSingleton.queries.SetExportDocStatus(c, db.SetExportDocStatusParams{
    ID: docID,
    Status: status,
  })
}

func SetExportStatus(c *gin.Context, exportID uuid.UUID, status db.ExportStatus) {
  // TODO: what to do if error happens here
  PrivacyServiceSingleton.queries.SetExportStatus(c, db.SetExportStatusParams{
    ID: exportID,
    Status: status,
  })
}

func UpdateExportStatus(err error, c *gin.Context, exportID uuid.UUID) {
  if err != nil {
    SetExportStatus(c, exportID, db.ExportStatusFailed)
  } else {
    SetExportStatus(c, exportID, db.ExportStatusComplete)
  }
}

func UpdateExportDocStatus(err error, c *gin.Context, docID uuid.UUID) {
  if err != nil {
    SetExportDocStatus(c, docID, db.ExportStatusFailed)
  } else {
    SetExportDocStatus(c, docID, db.ExportStatusComplete)
  }
}
