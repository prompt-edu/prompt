package courseMailing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/courseMailing/courseMailingDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type CourseMailingService struct {
	queries   db.Queries
	conn      *pgxpool.Pool
	clientURL string
}

var CourseMailingServiceSingleton *CourseMailingService

var (
	// ErrValidation signals invalid campaign input (maps to 400/422).
	ErrValidation = errors.New("course mailing validation error")
	// ErrNoRecipients signals that a send resolved zero recipients (maps to 422).
	ErrNoRecipients = errors.New("no recipients match the selected phase and statuses")
	// ErrSendInProgress signals a concurrent send (maps to 409).
	ErrSendInProgress = errors.New("a send is already in progress for this campaign")
	// ErrNotFound signals a missing campaign for the given course (maps to 404).
	ErrNotFound = errors.New("mail campaign not found")
)

func normalizeStatuses(statuses []string) []string {
	if statuses == nil {
		return []string{}
	}
	return statuses
}

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func marshalReplyTo(item *courseMailingDTO.MailItem) []byte {
	if item == nil {
		return nil
	}
	raw, err := json.Marshal(item)
	if err != nil {
		return nil
	}
	return raw
}

func marshalMailItems(items []courseMailingDTO.MailItem) []byte {
	if len(items) == 0 {
		return nil
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return nil
	}
	return raw
}

// expandStatuses turns the stored target statuses (which may include the "all"
// shortcut) into a concrete, deduplicated list of pass_status values.
func expandStatuses(raw []string) ([]string, error) {
	set := make(map[string]bool)
	for _, s := range raw {
		switch s {
		case "all":
			set[string(db.PassStatusPassed)] = true
			set[string(db.PassStatusFailed)] = true
			set[string(db.PassStatusNotAssessed)] = true
		case string(db.PassStatusPassed), string(db.PassStatusFailed), string(db.PassStatusNotAssessed):
			set[s] = true
		default:
			return nil, fmt.Errorf("%w: invalid target status %q", ErrValidation, s)
		}
	}
	if len(set) == 0 {
		return nil, fmt.Errorf("%w: at least one target status is required", ErrValidation)
	}
	out := make([]string, 0, len(set))
	for status := range set {
		out = append(out, status)
	}
	return out, nil
}

// withTx runs fn inside a transaction, rolling back on error and committing otherwise.
func (s *CourseMailingService) withTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// validatePhaseBelongsToCourse ensures a course phase is part of the given course,
// preventing a campaign from targeting a phase in an unrelated course.
func (s *CourseMailingService) validatePhaseBelongsToCourse(ctx context.Context, courseID, phaseID uuid.UUID) error {
	phaseCourseID, err := s.queries.GetCourseIDByCoursePhaseID(ctx, phaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: target course phase not found", ErrValidation)
	}
	if err != nil {
		return fmt.Errorf("failed to verify target course phase: %w", err)
	}
	if phaseCourseID != courseID {
		return fmt.Errorf("%w: target course phase does not belong to this course", ErrValidation)
	}
	return nil
}

// validateTargetPhase validates an optional target phase on create/update (drafts
// may leave it unset).
func (s *CourseMailingService) validateTargetPhase(ctx context.Context, courseID uuid.UUID, phaseID *uuid.UUID) error {
	if phaseID == nil {
		return nil
	}
	return s.validatePhaseBelongsToCourse(ctx, courseID, *phaseID)
}

// ensurePhaseInCourse resolves and validates the campaign's target phase against
// its own course. Used on every recipient-resolving / send path.
func (s *CourseMailingService) ensurePhaseInCourse(ctx context.Context, base db.MailCampaign) (uuid.UUID, error) {
	if !base.TargetCoursePhaseID.Valid {
		return uuid.UUID{}, fmt.Errorf("%w: a target course phase must be selected", ErrValidation)
	}
	phaseID := uuid.UUID(base.TargetCoursePhaseID.Bytes)
	if err := s.validatePhaseBelongsToCourse(ctx, base.CourseID, phaseID); err != nil {
		return uuid.UUID{}, err
	}
	return phaseID, nil
}

func (s *CourseMailingService) CreateCampaign(ctx context.Context, courseID uuid.UUID, actor courseMailingDTO.Actor, req courseMailingDTO.MailCampaignRequest) (db.MailCampaign, error) {
	if err := s.validateTargetPhase(ctx, courseID, req.TargetCoursePhaseID); err != nil {
		return db.MailCampaign{}, err
	}
	return s.queries.CreateMailCampaign(ctx, db.CreateMailCampaignParams{
		CourseID:            courseID,
		Name:                req.Name,
		Subject:             req.Subject,
		Body:                req.Body,
		TargetCoursePhaseID: toPgUUID(req.TargetCoursePhaseID),
		TargetPassStatuses:  normalizeStatuses(req.TargetPassStatuses),
		ReplyToOverride:     marshalReplyTo(req.ReplyToOverride),
		CcOverride:          marshalMailItems(req.CcOverride),
		BccOverride:         marshalMailItems(req.BccOverride),
		CreatedByID:         actor.ID,
		CreatedByEmail:      actor.Email,
		CreatedByName:       actor.Name,
		UpdatedByID:         actor.ID,
		UpdatedByEmail:      actor.Email,
		UpdatedByName:       actor.Name,
	})
}

func (s *CourseMailingService) getBase(ctx context.Context, courseID, campaignID uuid.UUID) (db.MailCampaign, error) {
	campaign, err := s.queries.GetMailCampaignBase(ctx, db.GetMailCampaignBaseParams{ID: campaignID, CourseID: courseID})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.MailCampaign{}, ErrNotFound
	}
	if err != nil {
		return db.MailCampaign{}, fmt.Errorf("failed to load mail campaign: %w", err)
	}
	return campaign, nil
}

func (s *CourseMailingService) UpdateCampaign(ctx context.Context, courseID, campaignID uuid.UUID, actor courseMailingDTO.Actor, req courseMailingDTO.MailCampaignRequest) (db.MailCampaign, error) {
	existing, err := s.getBase(ctx, courseID, campaignID)
	if err != nil {
		return db.MailCampaign{}, err
	}
	if existing.Status == db.MailCampaignStatusSending {
		return db.MailCampaign{}, ErrSendInProgress
	}
	if err := s.validateTargetPhase(ctx, courseID, req.TargetCoursePhaseID); err != nil {
		return db.MailCampaign{}, err
	}

	return s.queries.UpdateMailCampaign(ctx, db.UpdateMailCampaignParams{
		ID:                  campaignID,
		CourseID:            courseID,
		Name:                req.Name,
		Subject:             req.Subject,
		Body:                req.Body,
		TargetCoursePhaseID: toPgUUID(req.TargetCoursePhaseID),
		TargetPassStatuses:  normalizeStatuses(req.TargetPassStatuses),
		ReplyToOverride:     marshalReplyTo(req.ReplyToOverride),
		CcOverride:          marshalMailItems(req.CcOverride),
		BccOverride:         marshalMailItems(req.BccOverride),
		UpdatedByID:         actor.ID,
		UpdatedByEmail:      actor.Email,
		UpdatedByName:       actor.Name,
	})
}

func (s *CourseMailingService) ListCampaigns(ctx context.Context, courseID uuid.UUID) ([]courseMailingDTO.MailCampaign, error) {
	rows, err := s.queries.ListMailCampaignsForCourse(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list mail campaigns: %w", err)
	}
	campaigns := make([]courseMailingDTO.MailCampaign, 0, len(rows))
	for _, row := range rows {
		campaigns = append(campaigns, courseMailingDTO.MailCampaignFromListRow(row))
	}
	return campaigns, nil
}

func (s *CourseMailingService) GetCampaignDetail(ctx context.Context, courseID, campaignID uuid.UUID) (courseMailingDTO.MailCampaignDetail, error) {
	base, err := s.getBase(ctx, courseID, campaignID)
	if err != nil {
		return courseMailingDTO.MailCampaignDetail{}, err
	}

	recipientRows, err := s.queries.ListCampaignRecipientsWithStudent(ctx, campaignID)
	if err != nil {
		return courseMailingDTO.MailCampaignDetail{}, fmt.Errorf("failed to load campaign recipients: %w", err)
	}

	campaign := courseMailingDTO.MailCampaignFromModel(base)
	recipients := make([]courseMailingDTO.MailCampaignRecipient, 0, len(recipientRows))
	for _, row := range recipientRows {
		recipient := courseMailingDTO.RecipientFromRow(row)
		recipients = append(recipients, recipient)
		campaign.RecipientCount++
		switch row.Status {
		case db.MailRecipientStatusSent:
			campaign.SentCount++
		case db.MailRecipientStatusFailed:
			campaign.FailedCount++
		case db.MailRecipientStatusPending:
			campaign.PendingCount++
		}
	}

	return courseMailingDTO.MailCampaignDetail{MailCampaign: campaign, Recipients: recipients}, nil
}

func (s *CourseMailingService) CopyCampaign(ctx context.Context, courseID, campaignID uuid.UUID, actor courseMailingDTO.Actor) (db.MailCampaign, error) {
	base, err := s.getBase(ctx, courseID, campaignID)
	if err != nil {
		return db.MailCampaign{}, err
	}

	return s.queries.CreateMailCampaign(ctx, db.CreateMailCampaignParams{
		CourseID:            courseID,
		Name:                base.Name + " (Copy)",
		Subject:             base.Subject,
		Body:                base.Body,
		TargetCoursePhaseID: base.TargetCoursePhaseID,
		TargetPassStatuses:  normalizeStatuses(base.TargetPassStatuses),
		ReplyToOverride:     base.ReplyToOverride,
		CcOverride:          base.CcOverride,
		BccOverride:         base.BccOverride,
		CreatedByID:         actor.ID,
		CreatedByEmail:      actor.Email,
		CreatedByName:       actor.Name,
		UpdatedByID:         actor.ID,
		UpdatedByEmail:      actor.Email,
		UpdatedByName:       actor.Name,
	})
}

func (s *CourseMailingService) DeleteCampaign(ctx context.Context, courseID, campaignID uuid.UUID) error {
	base, err := s.getBase(ctx, courseID, campaignID)
	if err != nil {
		return err
	}
	if base.Status == db.MailCampaignStatusSending {
		return ErrSendInProgress
	}
	return s.queries.DeleteMailCampaign(ctx, db.DeleteMailCampaignParams{ID: campaignID, CourseID: courseID})
}

// resolveRecipients returns the live participant rows matching the campaign's
// target phase and statuses.
func (s *CourseMailingService) resolveRecipients(ctx context.Context, base db.MailCampaign) ([]db.GetParticipantMailingInformationForCampaignRow, error) {
	phaseID, err := s.ensurePhaseInCourse(ctx, base)
	if err != nil {
		return nil, err
	}
	statuses, err := expandStatuses(base.TargetPassStatuses)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.GetParticipantMailingInformationForCampaign(ctx, db.GetParticipantMailingInformationForCampaignParams{
		ID:      phaseID,
		Column2: statuses,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve campaign recipients: %w", err)
	}
	return rows, nil
}

func (s *CourseMailingService) PreviewRecipients(ctx context.Context, courseID, campaignID uuid.UUID) (courseMailingDTO.RecipientPreview, error) {
	base, err := s.getBase(ctx, courseID, campaignID)
	if err != nil {
		return courseMailingDTO.RecipientPreview{}, err
	}
	rows, err := s.resolveRecipients(ctx, base)
	if err != nil {
		return courseMailingDTO.RecipientPreview{}, err
	}
	items := make([]courseMailingDTO.RecipientPreviewItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, courseMailingDTO.RecipientPreviewItem{
			CourseParticipationID: row.CourseParticipationID,
			FirstName:             row.FirstName.String,
			LastName:              row.LastName.String,
			Email:                 row.Email.String,
		})
	}
	return courseMailingDTO.RecipientPreview{Count: len(items), Recipients: items}, nil
}

// ReconcileStuckCampaigns fails any campaigns left in the "sending" state by a
// prior crash/restart and finalizes their status. Called once on module init.
func (s *CourseMailingService) ReconcileStuckCampaigns(ctx context.Context) {
	ids, err := s.queries.GetSendingCampaignIDs(ctx)
	if err != nil {
		log.WithError(err).Warn("failed to load stuck mail campaigns for reconciliation")
		return
	}
	for _, id := range ids {
		if err := s.queries.FailStalePendingRecipients(ctx, db.FailStalePendingRecipientsParams{
			CampaignID:   id,
			ErrorMessage: "send interrupted by server restart",
		}); err != nil {
			log.WithError(err).WithField("campaignID", id).Warn("failed to fail stale recipients during reconciliation")
			continue
		}
		if err := s.finalizeCampaignStatus(ctx, id); err != nil {
			log.WithError(err).WithField("campaignID", id).Warn("failed to finalize stuck campaign status")
		}
	}
	if len(ids) > 0 {
		log.WithField("count", len(ids)).Info("reconciled stuck mail campaigns after restart")
	}
}

// finalizeCampaignStatus recomputes a campaign's status from its recipient rows
// (without touching the sent-by/sent-at metadata).
func (s *CourseMailingService) finalizeCampaignStatus(ctx context.Context, campaignID uuid.UUID) error {
	recipients, err := s.queries.ListCampaignRecipients(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to load recipients for status finalization: %w", err)
	}
	return s.queries.SetMailCampaignStatus(ctx, db.SetMailCampaignStatusParams{
		ID:     campaignID,
		Status: computeCampaignStatus(recipients),
	})
}

func computeCampaignStatus(recipients []db.MailCampaignRecipient) db.MailCampaignStatus {
	if len(recipients) == 0 {
		return db.MailCampaignStatusFailed
	}
	var sent, failed int
	for _, r := range recipients {
		switch r.Status {
		case db.MailRecipientStatusSent:
			sent++
		case db.MailRecipientStatusFailed:
			failed++
		}
	}
	switch {
	case failed == 0:
		return db.MailCampaignStatusSent
	case sent == 0:
		return db.MailCampaignStatusFailed
	default:
		return db.MailCampaignStatusPartiallyFailed
	}
}
