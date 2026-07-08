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
	ID                   uuid.UUID               `json:"id"`
	UserID               *uuid.UUID              `json:"userID"`
	StudentID            *uuid.UUID              `json:"studentID"`
	Status               db.ExportStatus         `json:"status"`
	DateCreated          time.Time               `json:"date_created"`
	ValidUntil           time.Time               `json:"valid_until"`
	NextRequestAllowedAt time.Time               `json:"next_request_allowed_at"`
	Documents            []PrivacyExportDocument `json:"documents"`
}

type AdminExportDoc struct {
	SourceName string          `json:"source_name"`
	Status     db.ExportStatus `json:"status"`
	Downloaded bool            `json:"downloaded"`
}

type AdminPrivacyExport struct {
	ID                   uuid.UUID        `json:"id"`
	UserID               *uuid.UUID       `json:"user_id"`
	StudentID            *uuid.UUID       `json:"student_id"`
	StudentFirstName     *string          `json:"student_first_name"`
	StudentLastName      *string          `json:"student_last_name"`
	StudentEmail         *string          `json:"student_email"`
	Status               db.ExportStatus  `json:"status"`
	DateCreated          time.Time        `json:"date_created"`
	ValidUntil           time.Time        `json:"valid_until"`
	NextRequestAllowedAt time.Time        `json:"next_request_allowed_at"`
	Docs                 []AdminExportDoc `json:"docs"`
}

func GetAdminPrivacyExportDTOFromDBModel(model db.GetAllExportsRow) (AdminPrivacyExport, error) {
	var docs []AdminExportDoc
	if err := json.Unmarshal(model.Docs, &docs); err != nil {
		return AdminPrivacyExport{}, fmt.Errorf("failed to parse export docs: %w", err)
	}

	return AdminPrivacyExport{
		ID:                   model.ID,
		UserID:               uuidPtr(model.UserID),
		StudentID:            uuidPtr(model.StudentID),
		StudentFirstName:     textPtr(model.StudentFirstName),
		StudentLastName:      textPtr(model.StudentLastName),
		StudentEmail:         textPtr(model.StudentEmail),
		Status:               model.Status,
		DateCreated:          model.DateCreated.Time,
		ValidUntil:           model.ValidUntil.Time,
		NextRequestAllowedAt: model.NextRequestAllowedAt.Time,
		Docs:                 docs,
	}, nil
}

func GetPrivacyExportDocDTOFromDBModel(model db.PrivacyExportDocument) PrivacyExportDocument {
	return PrivacyExportDocument{
		ID:           model.ID,
		DateCreated:  model.DateCreated.Time,
		SourceName:   model.SourceName,
		Status:       model.Status,
		DownloadedAt: timePtr(model.DownloadedAt),
	}
}

func GetPrivacyExportDTOFromDBModel(model db.PrivacyExport) PrivacyExport {
	return PrivacyExport{
		ID:                   model.ID,
		UserID:               uuidPtr(model.UserID),
		StudentID:            uuidPtr(model.StudentID),
		Status:               model.Status,
		DateCreated:          model.DateCreated.Time,
		ValidUntil:           model.ValidUntil.Time,
		NextRequestAllowedAt: model.NextRequestAllowedAt.Time,
		Documents:            []PrivacyExportDocument{},
	}
}

func GetPrivacyExportWithDocsDTOFromDBModel(model db.PrivacyExportWithDoc) (PrivacyExport, error) {
	var documents []PrivacyExportDocument
	if err := json.Unmarshal(model.Documents, &documents); err != nil {
		return PrivacyExport{}, fmt.Errorf("failed to parse export documents: %w", err)
	}

	return PrivacyExport{
		ID:                   model.ID,
		UserID:               uuidPtr(model.UserID),
		StudentID:            uuidPtr(model.StudentID),
		Status:               model.Status,
		DateCreated:          model.DateCreated.Time,
		ValidUntil:           model.ValidUntil.Time,
		NextRequestAllowedAt: model.NextRequestAllowedAt.Time,
		Documents:            documents,
	}, nil
}
