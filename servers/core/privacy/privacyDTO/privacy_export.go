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
	ObjectKey    string          `json:"object_key"`
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

func GetPrivacyExportDocDTOFromDBModel(model db.PrivacyExportDocument) PrivacyExportDocument {
	dto := PrivacyExportDocument{
		ID:          model.ID,
		DateCreated: model.DateCreated.Time,
		SourceName:  model.SourceName,
		ObjectKey:   model.ObjectKey,
		Status:      model.Status,
	}
	if model.DownloadedAt.Valid {
		dto.DownloadedAt = &model.DownloadedAt.Time
	}
	return dto
}

func GetPrivacyExportDTOFromDBModel(model db.PrivacyExport) PrivacyExport {
	var studentID *uuid.UUID
	if model.Studentid.Valid {
		id, _ := uuid.FromBytes(model.Studentid.Bytes[:])
		studentID = &id
	}

	return PrivacyExport{
		ID:          model.ID,
		UserID:      model.Userid,
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
		UserID:      model.Userid,
		Status:      model.Status,
		DateCreated: model.DateCreated.Time,
		ValidUntil:  model.ValidUntil.Time,
	}

	if model.Studentid.Valid {
		id, _ := uuid.FromBytes(model.Studentid.Bytes[:])
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
