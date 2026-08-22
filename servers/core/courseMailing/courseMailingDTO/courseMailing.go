package courseMailingDTO

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// MailItem is a named email address (reply-to / cc / bcc entry).
type MailItem struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Actor identifies the Keycloak user who performed an action on a campaign.
type Actor struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// MailCampaignRequest is the create/update payload sent by the client.
type MailCampaignRequest struct {
	Name                string     `json:"name"`
	Subject             string     `json:"subject"`
	Body                string     `json:"body"`
	TargetCoursePhaseID *uuid.UUID `json:"targetCoursePhaseID"`
	TargetPassStatuses  []string   `json:"targetPassStatuses"`
	ReplyToOverride     *MailItem  `json:"replyToOverride"`
	CcOverride          []MailItem `json:"ccOverride"`
	BccOverride         []MailItem `json:"bccOverride"`
}

// MailCampaign is the API representation of a campaign, including recipient counts.
type MailCampaign struct {
	ID                  uuid.UUID  `json:"id"`
	CourseID            uuid.UUID  `json:"courseID"`
	Name                string     `json:"name"`
	Subject             string     `json:"subject"`
	Body                string     `json:"body"`
	TargetCoursePhaseID *uuid.UUID `json:"targetCoursePhaseID"`
	TargetPassStatuses  []string   `json:"targetPassStatuses"`
	ReplyToOverride     *MailItem  `json:"replyToOverride"`
	CcOverride          []MailItem `json:"ccOverride"`
	BccOverride         []MailItem `json:"bccOverride"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"createdAt"`
	CreatedBy           Actor      `json:"createdBy"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	UpdatedBy           Actor      `json:"updatedBy"`
	SentAt              *time.Time `json:"sentAt"`
	SentBy              *Actor     `json:"sentBy"`
	RecipientCount      int64      `json:"recipientCount"`
	SentCount           int64      `json:"sentCount"`
	FailedCount         int64      `json:"failedCount"`
	PendingCount        int64      `json:"pendingCount"`
}

// MailCampaignDetail adds the resolved/snapshotted recipient list.
type MailCampaignDetail struct {
	MailCampaign
	Recipients []MailCampaignRecipient `json:"recipients"`
}

// MailCampaignRecipient is a single recipient row with its send outcome.
type MailCampaignRecipient struct {
	CourseParticipationID uuid.UUID  `json:"courseParticipationID"`
	FirstName             string     `json:"firstName"`
	LastName              string     `json:"lastName"`
	Email                 string     `json:"email"`
	Status                string     `json:"status"`
	ErrorMessage          string     `json:"errorMessage"`
	SentAt                *time.Time `json:"sentAt"`
}

// RecipientPreviewItem is a live-resolved recipient (before sending).
type RecipientPreviewItem struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	FirstName             string    `json:"firstName"`
	LastName              string    `json:"lastName"`
	Email                 string    `json:"email"`
}

// RecipientPreview is the response for the recipient preview endpoint.
type RecipientPreview struct {
	Count      int                    `json:"count"`
	Recipients []RecipientPreviewItem `json:"recipients"`
}

// SendResponse is returned when a send is accepted for async processing.
type SendResponse struct {
	RecipientCount int `json:"recipientCount"`
}

// --- converters -------------------------------------------------------------

func parseMailItem(raw []byte) *MailItem {
	if len(raw) == 0 {
		return nil
	}
	var item MailItem
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil
	}
	if item.Email == "" {
		return nil
	}
	return &item
}

func parseMailItems(raw []byte) []MailItem {
	if len(raw) == 0 {
		return nil
	}
	var items []MailItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}

func pgUUIDToPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	parsed := uuid.UUID(id.Bytes)
	return &parsed
}

func pgTimeToPtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func sentActor(id, email, name pgtype.Text) *Actor {
	if !id.Valid && !email.Valid && !name.Valid {
		return nil
	}
	return &Actor{ID: id.String, Email: email.String, Name: name.String}
}

// MailCampaignFromListRow maps a list row (with counts) to the API DTO.
func MailCampaignFromListRow(row db.ListMailCampaignsForCourseRow) MailCampaign {
	return MailCampaign{
		ID:                  row.ID,
		CourseID:            row.CourseID,
		Name:                row.Name,
		Subject:             row.Subject,
		Body:                row.Body,
		TargetCoursePhaseID: pgUUIDToPtr(row.TargetCoursePhaseID),
		TargetPassStatuses:  row.TargetPassStatuses,
		ReplyToOverride:     parseMailItem(row.ReplyToOverride),
		CcOverride:          parseMailItems(row.CcOverride),
		BccOverride:         parseMailItems(row.BccOverride),
		Status:              string(row.Status),
		CreatedAt:           row.CreatedAt.Time,
		CreatedBy:           Actor{ID: row.CreatedByID, Email: row.CreatedByEmail, Name: row.CreatedByName},
		UpdatedAt:           row.UpdatedAt.Time,
		UpdatedBy:           Actor{ID: row.UpdatedByID, Email: row.UpdatedByEmail, Name: row.UpdatedByName},
		SentAt:              pgTimeToPtr(row.SentAt),
		SentBy:              sentActor(row.SentByID, row.SentByEmail, row.SentByName),
		RecipientCount:      row.RecipientCount,
		SentCount:           row.SentCount,
		FailedCount:         row.FailedCount,
		PendingCount:        row.PendingCount,
	}
}

// MailCampaignFromModel maps a base campaign row to the API DTO (without counts).
func MailCampaignFromModel(m db.MailCampaign) MailCampaign {
	return MailCampaign{
		ID:                  m.ID,
		CourseID:            m.CourseID,
		Name:                m.Name,
		Subject:             m.Subject,
		Body:                m.Body,
		TargetCoursePhaseID: pgUUIDToPtr(m.TargetCoursePhaseID),
		TargetPassStatuses:  m.TargetPassStatuses,
		ReplyToOverride:     parseMailItem(m.ReplyToOverride),
		CcOverride:          parseMailItems(m.CcOverride),
		BccOverride:         parseMailItems(m.BccOverride),
		Status:              string(m.Status),
		CreatedAt:           m.CreatedAt.Time,
		CreatedBy:           Actor{ID: m.CreatedByID, Email: m.CreatedByEmail, Name: m.CreatedByName},
		UpdatedAt:           m.UpdatedAt.Time,
		UpdatedBy:           Actor{ID: m.UpdatedByID, Email: m.UpdatedByEmail, Name: m.UpdatedByName},
		SentAt:              pgTimeToPtr(m.SentAt),
		SentBy:              sentActor(m.SentByID, m.SentByEmail, m.SentByName),
	}
}

// RecipientFromRow maps a stored recipient row (with student name) to the API DTO.
func RecipientFromRow(row db.ListCampaignRecipientsWithStudentRow) MailCampaignRecipient {
	return MailCampaignRecipient{
		CourseParticipationID: row.CourseParticipationID,
		FirstName:             row.FirstName.String,
		LastName:              row.LastName.String,
		Email:                 row.Email,
		Status:                string(row.Status),
		ErrorMessage:          row.ErrorMessage,
		SentAt:                pgTimeToPtr(row.SentAt),
	}
}
