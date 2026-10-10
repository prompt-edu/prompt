package announcement

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/prompt-edu/prompt/servers/core/announcement/announcementDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type AnnouncementService struct {
	queries db.Queries
}

func NewAnnouncementService(queries db.Queries) *AnnouncementService {
	return &AnnouncementService{queries: queries}
}

func (s *AnnouncementService) ListActiveAnnouncements(ctx context.Context) ([]announcementDTO.Announcement, error) {
	announcements, err := s.queries.ListActiveAnnouncements(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active announcements: %w", err)
	}
	return announcementDTO.AnnouncementsFromDBModels(announcements), nil
}

func (s *AnnouncementService) ListAnnouncements(ctx context.Context, includeExpired bool) ([]announcementDTO.Announcement, error) {
	announcements, err := s.queries.ListAnnouncements(ctx, includeExpired)
	if err != nil {
		return nil, fmt.Errorf("failed to list announcements: %w", err)
	}
	return announcementDTO.AnnouncementsFromDBModels(announcements), nil
}

func (s *AnnouncementService) CreateAnnouncement(ctx context.Context, request announcementDTO.UpsertAnnouncement) (announcementDTO.Announcement, error) {
	created, err := s.queries.CreateAnnouncement(ctx, db.CreateAnnouncementParams{
		Severity:  db.AnnouncementSeverity(request.Severity),
		Title:     request.Title,
		Message:   request.Message,
		LinkUrl:   request.LinkURL,
		LinkLabel: request.LinkLabel,
		StartsAt:  announcementDTO.ToTimestamptz(request.StartsAt),
		ExpiresAt: announcementDTO.ToTimestamptz(request.ExpiresAt),
		Enabled:   request.Enabled,
	})
	if err != nil {
		return announcementDTO.Announcement{}, fmt.Errorf("failed to create announcement: %w", err)
	}
	return announcementDTO.AnnouncementFromDBModel(created), nil
}

func (s *AnnouncementService) UpdateAnnouncement(ctx context.Context, id uuid.UUID, request announcementDTO.UpsertAnnouncement) (announcementDTO.Announcement, error) {
	updated, err := s.queries.UpdateAnnouncement(ctx, db.UpdateAnnouncementParams{
		ID:        id,
		Severity:  db.AnnouncementSeverity(request.Severity),
		Title:     request.Title,
		Message:   request.Message,
		LinkUrl:   request.LinkURL,
		LinkLabel: request.LinkLabel,
		StartsAt:  announcementDTO.ToTimestamptz(request.StartsAt),
		ExpiresAt: announcementDTO.ToTimestamptz(request.ExpiresAt),
		Enabled:   request.Enabled,
	})
	if err != nil {
		return announcementDTO.Announcement{}, fmt.Errorf("failed to update announcement %s: %w", id, err)
	}
	return announcementDTO.AnnouncementFromDBModel(updated), nil
}

func (s *AnnouncementService) DeleteAnnouncement(ctx context.Context, id uuid.UUID) error {
	deleted, err := s.queries.DeleteAnnouncement(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete announcement %s: %w", id, err)
	}
	if deleted == 0 {
		return fmt.Errorf("announcement %s: %w", id, pgx.ErrNoRows)
	}
	return nil
}
