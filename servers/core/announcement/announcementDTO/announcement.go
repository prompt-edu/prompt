package announcementDTO

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type Announcement struct {
	ID        uuid.UUID  `json:"id"`
	Severity  string     `json:"severity"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	LinkURL   string     `json:"linkUrl"`
	LinkLabel string     `json:"linkLabel"`
	StartsAt  *time.Time `json:"startsAt"`
	ExpiresAt *time.Time `json:"expiresAt"`
	Enabled   bool       `json:"enabled"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type UpsertAnnouncement struct {
	Severity  string     `json:"severity"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	LinkURL   string     `json:"linkUrl"`
	LinkLabel string     `json:"linkLabel"`
	StartsAt  *time.Time `json:"startsAt"`
	ExpiresAt *time.Time `json:"expiresAt"`
	Enabled   bool       `json:"enabled"`
}

func AnnouncementFromDBModel(model db.Announcement) Announcement {
	return Announcement{
		ID:        model.ID,
		Severity:  string(model.Severity),
		Title:     model.Title,
		Message:   model.Message,
		LinkURL:   model.LinkUrl,
		LinkLabel: model.LinkLabel,
		StartsAt:  optionalTime(model.StartsAt),
		ExpiresAt: optionalTime(model.ExpiresAt),
		Enabled:   model.Enabled,
		CreatedAt: model.CreatedAt.Time,
		UpdatedAt: model.UpdatedAt.Time,
	}
}

func AnnouncementsFromDBModels(models []db.Announcement) []Announcement {
	announcements := make([]Announcement, 0, len(models))
	for _, m := range models {
		announcements = append(announcements, AnnouncementFromDBModel(m))
	}
	return announcements
}

func ToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func optionalTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}
