package privacyDTO

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type PrivacyExportDocument struct {
	ID           uuid.UUID       `json:"id"`
	DateCreated  time.Time       `json:"date_created"`
	SourceName   string          `json:"source_name"`
	Status       db.ExportStatus `json:"status"`
	FileSize     *int64          `json:"file_size"`
	DownloadedAt *time.Time      `json:"downloaded_at"`
}

type PrivacyExport struct {
	ID             uuid.UUID               `json:"id"`
	UserID         uuid.UUID               `json:"userID"`
	StudentID      *uuid.UUID              `json:"studentID"`
	Status         db.ExportStatus         `json:"status"`
	DateCreated    time.Time               `json:"date_created"`
	ValidUntil     time.Time               `json:"valid_until"`
	Documents      []PrivacyExportDocument `json:"documents"`
}

type AdminPrivacyExport struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"userID"`
	StudentID        *uuid.UUID      `json:"studentID"`
	Status           db.ExportStatus `json:"status"`
	DateCreated      time.Time       `json:"date_created"`
	ValidUntil       time.Time       `json:"valid_until"`
	TotalDocs        int32           `json:"total_docs"`
	DownloadedDocs   int32           `json:"downloaded_docs"`
	LastDownloadedAt *time.Time      `json:"last_downloaded_at"`
	FailedDocs       []string        `json:"failed_docs"`
}

func GetAdminPrivacyExportDTOFromDBModel(model db.GetAllExportsRow) AdminPrivacyExport {
	var studentID *uuid.UUID
	if model.StudentID.Valid {
		id, _ := uuid.FromBytes(model.StudentID.Bytes[:])
		studentID = &id
	}

	failedDocs := []string{}
	if arr, ok := model.FailedDocs.([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				failedDocs = append(failedDocs, s)
			}
		}
	}

	dto := AdminPrivacyExport{
		ID:             model.ID,
		UserID:         model.UserID,
		StudentID:      studentID,
		Status:         model.Status,
		DateCreated:    model.DateCreated.Time,
		ValidUntil:     model.ValidUntil.Time,
		TotalDocs:      model.TotalDocs,
		DownloadedDocs: model.DownloadedDocs,
		FailedDocs:     failedDocs,
	}
	if model.LastDownloadedAt.Valid {
		dto.LastDownloadedAt = &model.LastDownloadedAt.Time
	}
	return dto
}

func GetPrivacyExportDocDTOFromDBModel(model db.PrivacyExportDocument) PrivacyExportDocument {
	dto := PrivacyExportDocument{
		ID:          model.ID,
		DateCreated: model.DateCreated.Time,
		SourceName:  model.SourceName,
		Status:      model.Status,
	}
	if model.DownloadedAt.Valid {
		dto.DownloadedAt = &model.DownloadedAt.Time
	}
	return dto
}

func GetPrivacyExportDTOFromDBModel(model db.PrivacyExport) PrivacyExport {
	var studentID *uuid.UUID
	if model.StudentID.Valid {
		id, _ := uuid.FromBytes(model.StudentID.Bytes[:])
		studentID = &id
	}

	return PrivacyExport{
		ID:          model.ID,
		UserID:      model.UserID,
		StudentID:   studentID,
		Status:      model.Status,
		DateCreated: model.DateCreated.Time,
    ValidUntil:  model.ValidUntil.Time,
		Documents:   []PrivacyExportDocument{},
	}
}

func GetPrivacyExportWithDocsDTOFromDBModel(model db.PrivacyExportWithDoc) (PrivacyExport, error) {
	export := PrivacyExport{
		ID:          model.ID,
		UserID:      model.UserID,
		Status:      model.Status,
		DateCreated: model.DateCreated.Time,
		ValidUntil:  model.ValidUntil.Time,
	}

	if model.StudentID.Valid {
		id, _ := uuid.FromBytes(model.StudentID.Bytes[:])
		export.StudentID = &id
	}

	rawDocs, err := json.Marshal(model.Documents)
	if err != nil {
		return PrivacyExport{}, fmt.Errorf("failed to serialize documents field: %w", err)
	}
	var documents []PrivacyExportDocument
	if err := json.Unmarshal(rawDocs, &documents); err != nil {
		return PrivacyExport{}, fmt.Errorf("failed to parse export documents: %w", err)
	}
	export.Documents = documents

	return export, nil
}
