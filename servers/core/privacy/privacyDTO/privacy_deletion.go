package privacyDTO

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type AuditorDecision string

const (
	AuditorDecisionApprove AuditorDecision = "approve"
	AuditorDecisionReject  AuditorDecision = "reject"
)

type AuditorDecisionRequest struct {
	Decision AuditorDecision `json:"decision" binding:"required,oneof=approve reject"`
	Note     string          `json:"note"`
}

type AdminInitiateDeletionBody struct {
	StudentIDs []uuid.UUID `json:"student_ids" binding:"required,min=1,dive,required"`
}

type AdminPrivacyDeletionRequest struct {
	ID                 uuid.UUID                       `json:"id"`
	UserID             *uuid.UUID                      `json:"user_id"`
	StudentID          *uuid.UUID                      `json:"student_id"`
	StudentFirstName   *string                         `json:"student_first_name"`
	StudentLastName    *string                         `json:"student_last_name"`
	StudentEmail       *string                         `json:"student_email"`
	Status             db.PrivacyDeletionRequestStatus `json:"status"`
	RequestedAt        time.Time                       `json:"requested_at"`
	AuditorID          *uuid.UUID                      `json:"auditor_id"`
	AuditorName        string                          `json:"auditor_name"`
	AuditorEmail       string                          `json:"auditor_email"`
	AuditorRespondedAt *time.Time                      `json:"auditor_responded_at"`
	AuditorNote        string                          `json:"auditor_note"`
	CompletedAt        *time.Time                      `json:"completed_at"`
	Subrequests        []PrivacyDeletionSubrequest     `json:"subrequests"`
}

func GetAdminPrivacyDeletionRequestDTOFromDBModel(model db.GetAllDeletionRequestsRow) (AdminPrivacyDeletionRequest, error) {
	var subrequests []PrivacyDeletionSubrequest
	if err := json.Unmarshal(model.Subrequests, &subrequests); err != nil {
		return AdminPrivacyDeletionRequest{}, fmt.Errorf("failed to parse subrequests: %w", err)
	}

	return AdminPrivacyDeletionRequest{
		ID:                 model.ID,
		UserID:             uuidPtr(model.UserID),
		StudentID:          uuidPtr(model.StudentID),
		StudentFirstName:   textPtr(model.StudentFirstName),
		StudentLastName:    textPtr(model.StudentLastName),
		StudentEmail:       textPtr(model.StudentEmail),
		Status:             model.Status,
		RequestedAt:        model.RequestedAt.Time,
		AuditorID:          uuidPtr(model.AuditorID),
		AuditorName:        model.AuditorName,
		AuditorEmail:       model.AuditorEmail,
		AuditorRespondedAt: timePtr(model.AuditorRespondedAt),
		AuditorNote:        model.AuditorNote,
		CompletedAt:        timePtr(model.CompletedAt),
		Subrequests:        subrequests,
	}, nil
}

type PrivacyDeletionSubrequest struct {
	ID           uuid.UUID                          `json:"id"`
	SourceName   string                             `json:"source_name"`
	Status       db.PrivacyDeletionSubrequestStatus `json:"status"`
	CreatedAt    time.Time                          `json:"created_at"`
	CompletedAt  *time.Time                         `json:"completed_at"`
	ErrorMessage string                             `json:"error_message"`
}

// PublicPrivacyDeletionSubrequest is the requester-facing subrequest view.
// ErrorMessage is intentionally omitted so technical failure details never reach end users.
type PublicPrivacyDeletionSubrequest struct {
	ID          uuid.UUID                          `json:"id"`
	SourceName  string                             `json:"source_name"`
	Status      db.PrivacyDeletionSubrequestStatus `json:"status"`
	CreatedAt   time.Time                          `json:"created_at"`
	CompletedAt *time.Time                         `json:"completed_at"`
}

type PrivacyDeletionRequest struct {
	ID                 uuid.UUID                         `json:"id"`
	UserID             *uuid.UUID                        `json:"user_id"`
	StudentID          *uuid.UUID                        `json:"student_id"`
	RequestedAt        time.Time                         `json:"requested_at"`
	Status             db.PrivacyDeletionRequestStatus   `json:"status"`
	AuditorID          *uuid.UUID                        `json:"auditor_id"`
	AuditorName        string                            `json:"auditor_name"`
	AuditorEmail       string                            `json:"auditor_email"`
	AuditorRespondedAt *time.Time                        `json:"auditor_responded_at"`
	AuditorNote        string                            `json:"auditor_note"`
	CompletedAt        *time.Time                        `json:"completed_at"`
	Subrequests        []PublicPrivacyDeletionSubrequest `json:"subrequests"`
}

func GetPrivacyDeletionSubrequestDTOFromDBModel(model db.PrivacyDeletionSubrequest) PrivacyDeletionSubrequest {
	return PrivacyDeletionSubrequest{
		ID:           model.ID,
		SourceName:   model.SourceName,
		Status:       model.Status,
		CreatedAt:    model.CreatedAt.Time,
		CompletedAt:  timePtr(model.CompletedAt),
		ErrorMessage: model.ErrorMessage,
	}
}

func GetPrivacyDeletionRequestDTOFromDBModel(model db.PrivacyDeletionRequest) PrivacyDeletionRequest {
	return PrivacyDeletionRequest{
		ID:                 model.ID,
		UserID:             uuidPtr(model.UserID),
		StudentID:          uuidPtr(model.StudentID),
		RequestedAt:        model.RequestedAt.Time,
		Status:             model.Status,
		AuditorID:          uuidPtr(model.AuditorID),
		AuditorName:        model.AuditorName,
		AuditorEmail:       model.AuditorEmail,
		AuditorRespondedAt: timePtr(model.AuditorRespondedAt),
		AuditorNote:        model.AuditorNote,
		CompletedAt:        timePtr(model.CompletedAt),
		Subrequests:        []PublicPrivacyDeletionSubrequest{},
	}
}

func GetPrivacyDeletionRequestWithSubrequestsDTOFromDBModel(model db.PrivacyDeletionRequestWithSubrequest) (PrivacyDeletionRequest, error) {
	var subrequests []PublicPrivacyDeletionSubrequest
	if err := json.Unmarshal(model.Subrequests, &subrequests); err != nil {
		return PrivacyDeletionRequest{}, fmt.Errorf("failed to parse subrequests: %w", err)
	}

	return PrivacyDeletionRequest{
		ID:                 model.ID,
		UserID:             uuidPtr(model.UserID),
		StudentID:          uuidPtr(model.StudentID),
		RequestedAt:        model.RequestedAt.Time,
		Status:             model.Status,
		AuditorID:          uuidPtr(model.AuditorID),
		AuditorName:        model.AuditorName,
		AuditorEmail:       model.AuditorEmail,
		AuditorRespondedAt: timePtr(model.AuditorRespondedAt),
		AuditorNote:        model.AuditorNote,
		CompletedAt:        timePtr(model.CompletedAt),
		Subrequests:        subrequests,
	}, nil
}
