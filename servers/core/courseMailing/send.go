package courseMailing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/core/courseMailing/courseMailingDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	log "github.com/sirupsen/logrus"
)

const maxConcurrentSends = 8

// campaignSendTimeout bounds how long a single campaign send may run before the
// dispatch loop stops launching new sends. A var (not const) so tests can shrink it.
var campaignSendTimeout = 30 * time.Minute

// sendMailFn is the SMTP send entry point, swappable in tests.
var sendMailFn = mailing.SendCourseMail

var nowFn = func() time.Time { return time.Now().UTC() }

// runSendAsync dispatches the campaign send. Overridden in tests to run synchronously.
var runSendAsync = func(fn func()) { go fn() }

type campaignSendItem struct {
	recipientID uuid.UUID
	email       string
	subject     string
	body        string
}

type campaignSendJob struct {
	campaignID  uuid.UUID
	subject     string // raw template, for the archive copy
	body        string // raw template, for the archive copy
	link        string // resolved {{coursePhaseLink}}, for the archive copy
	items       []campaignSendItem
	settings    mailingDTO.CourseMailingSettings
	courseInfo  db.GetPassedMailingInformationRow
	actor       courseMailingDTO.Actor
	setSentMeta bool
}

// SendCampaign validates, resolves recipients live, snapshots them, and dispatches
// the send in a detached goroutine. Returns the recipient count.
func (s *CourseMailingService) SendCampaign(ctx context.Context, courseID, campaignID uuid.UUID, actor courseMailingDTO.Actor) (int, error) {
	base, err := s.getBase(ctx, courseID, campaignID)
	if err != nil {
		return 0, err
	}
	switch base.Status {
	case db.MailCampaignStatusSent, db.MailCampaignStatusPartiallyFailed, db.MailCampaignStatusFailed:
		return 0, fmt.Errorf("%w: campaign has already been sent; use resend failed or copy the campaign to send again", ErrValidation)
	}
	if err := validateCampaignForSend(base); err != nil {
		return 0, err
	}

	phaseID, err := s.ensurePhaseInCourse(ctx, base)
	if err != nil {
		return 0, err
	}
	settings, err := s.resolveSettings(ctx, base, phaseID)
	if err != nil {
		return 0, err
	}
	courseInfo, err := s.queries.GetPassedMailingInformation(ctx, phaseID)
	if err != nil {
		return 0, fmt.Errorf("failed to load course mailing context: %w", err)
	}

	rows, err := s.resolveRecipients(ctx, base)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, ErrNoRecipients
	}

	// Atomically flip to "sending" and snapshot recipients in one transaction, so a
	// failed snapshot never leaves the campaign stuck in the sending state.
	if err := s.withTx(ctx, func(q *db.Queries) error {
		if err := s.trySetCampaignSending(ctx, q, courseID, campaignID); err != nil {
			return err
		}
		if err := q.DeleteCampaignRecipients(ctx, campaignID); err != nil {
			return fmt.Errorf("failed to reset campaign recipients: %w", err)
		}
		for _, row := range rows {
			if err := q.InsertCampaignRecipient(ctx, db.InsertCampaignRecipientParams{
				CampaignID:            campaignID,
				CourseID:              courseID,
				CourseParticipationID: row.CourseParticipationID,
				Email:                 row.Email.String,
			}); err != nil {
				return fmt.Errorf("failed to snapshot campaign recipient: %w", err)
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, ErrSendInProgress) {
			return 0, ErrSendInProgress
		}
		return 0, err
	}

	recipientIDByCP, err := s.recipientIDMap(ctx, campaignID)
	if err != nil {
		return 0, err
	}

	link := s.buildCoursePhaseLink(ctx, phaseID)
	items := make([]campaignSendItem, 0, len(rows))
	for _, row := range rows {
		recipientID, ok := recipientIDByCP[row.CourseParticipationID]
		if !ok {
			continue
		}
		items = append(items, s.buildSendItem(recipientID, row.Email.String, base, courseInfo, toMailingRow(row), link))
	}

	job := campaignSendJob{
		campaignID:  campaignID,
		subject:     base.Subject,
		body:        base.Body,
		link:        link,
		items:       items,
		settings:    settings,
		courseInfo:  courseInfo,
		actor:       actor,
		setSentMeta: true,
	}
	runSendAsync(func() { s.runCampaignSend(job) })

	return len(items), nil
}

// ResendFailed re-sends only the recipients that previously failed, re-resolving
// their placeholder data live. The original sent metadata is preserved.
func (s *CourseMailingService) ResendFailed(ctx context.Context, courseID, campaignID uuid.UUID, actor courseMailingDTO.Actor) (int, error) {
	base, err := s.getBase(ctx, courseID, campaignID)
	if err != nil {
		return 0, err
	}
	if base.Status == db.MailCampaignStatusSending {
		return 0, ErrSendInProgress
	}
	if err := validateCampaignForSend(base); err != nil {
		return 0, err
	}

	failed, err := s.queries.ListFailedCampaignRecipients(ctx, campaignID)
	if err != nil {
		return 0, fmt.Errorf("failed to load failed recipients: %w", err)
	}
	if len(failed) == 0 {
		return 0, fmt.Errorf("%w: no failed recipients to resend", ErrValidation)
	}

	phaseID, err := s.ensurePhaseInCourse(ctx, base)
	if err != nil {
		return 0, err
	}
	settings, err := s.resolveSettings(ctx, base, phaseID)
	if err != nil {
		return 0, err
	}
	courseInfo, err := s.queries.GetPassedMailingInformation(ctx, phaseID)
	if err != nil {
		return 0, fmt.Errorf("failed to load course mailing context: %w", err)
	}

	statuses, err := expandStatuses(base.TargetPassStatuses)
	if err != nil {
		return 0, err
	}

	cpIDs := make([]uuid.UUID, 0, len(failed))
	failedRecipientID := make(map[uuid.UUID]uuid.UUID, len(failed))
	for _, r := range failed {
		cpIDs = append(cpIDs, r.CourseParticipationID)
		failedRecipientID[r.CourseParticipationID] = r.ID
	}
	liveRows, err := s.queries.GetCampaignRecipientMailingInfoByIDs(ctx, db.GetCampaignRecipientMailingInfoByIDsParams{ID: phaseID, Column2: cpIDs, Column3: statuses})
	if err != nil {
		return 0, fmt.Errorf("failed to re-resolve failed recipients: %w", err)
	}
	if len(liveRows) == 0 {
		// None of the failed recipients are valid participants anymore; leave the
		// campaign untouched (do not flip it to sending).
		return 0, fmt.Errorf("%w: failed recipients are no longer valid participants", ErrValidation)
	}

	// Atomically guard + reset the resolvable failed recipients to pending.
	if err := s.withTx(ctx, func(q *db.Queries) error {
		if err := s.trySetCampaignSending(ctx, q, courseID, campaignID); err != nil {
			return err
		}
		for _, row := range liveRows {
			if err := q.InsertCampaignRecipient(ctx, db.InsertCampaignRecipientParams{
				CampaignID:            campaignID,
				CourseID:              courseID,
				CourseParticipationID: row.CourseParticipationID,
				Email:                 row.Email.String,
			}); err != nil {
				return fmt.Errorf("failed to reset failed recipient: %w", err)
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, ErrSendInProgress) {
			return 0, ErrSendInProgress
		}
		return 0, err
	}

	link := s.buildCoursePhaseLink(ctx, phaseID)
	items := make([]campaignSendItem, 0, len(liveRows))
	for _, row := range liveRows {
		recipientID, ok := failedRecipientID[row.CourseParticipationID]
		if !ok {
			continue
		}
		mailingRow := db.GetParticipantMailingInformationRow{
			FirstName:           row.FirstName,
			LastName:            row.LastName,
			Email:               row.Email,
			MatriculationNumber: row.MatriculationNumber,
			UniversityLogin:     row.UniversityLogin,
			StudyDegree:         row.StudyDegree,
			CurrentSemester:     row.CurrentSemester,
			StudyProgram:        row.StudyProgram,
		}
		items = append(items, s.buildSendItem(recipientID, row.Email.String, base, courseInfo, mailingRow, link))
	}

	job := campaignSendJob{
		campaignID:  campaignID,
		subject:     base.Subject,
		body:        base.Body,
		link:        link,
		items:       items,
		settings:    settings,
		courseInfo:  courseInfo,
		actor:       actor,
		setSentMeta: false,
	}
	runSendAsync(func() { s.runCampaignSend(job) })

	return len(items), nil
}

// TestSend sends a single rendered copy of the campaign to the acting user's own
// email address, using a sample recipient's placeholder values when available.
func (s *CourseMailingService) TestSend(ctx context.Context, courseID, campaignID uuid.UUID, actor courseMailingDTO.Actor) error {
	base, err := s.getBase(ctx, courseID, campaignID)
	if err != nil {
		return err
	}
	if err := validateCampaignForTest(base); err != nil {
		return err
	}
	if actor.Email == "" {
		return fmt.Errorf("%w: your account has no email address for the test send", ErrValidation)
	}

	phaseID, err := s.ensurePhaseInCourse(ctx, base)
	if err != nil {
		return err
	}
	settings, err := s.resolveSettings(ctx, base, phaseID)
	if err != nil {
		return err
	}
	courseInfo, err := s.queries.GetPassedMailingInformation(ctx, phaseID)
	if err != nil {
		return fmt.Errorf("failed to load course mailing context: %w", err)
	}

	sample := db.GetParticipantMailingInformationRow{}
	if rows, err := s.resolveRecipients(ctx, base); err == nil && len(rows) > 0 {
		sample = toMailingRow(rows[0])
	}

	placeholders := mailing.StatusPlaceholderValues(courseInfo.CourseName, courseInfo.CourseStartDate, courseInfo.CourseEndDate, sample)
	if link := s.buildCoursePhaseLink(ctx, phaseID); link != "" {
		placeholders["coursePhaseLink"] = link
	}

	subject := "[TEST] " + mailing.ReplacePlaceholders(base.Subject, placeholders)
	body := mailing.ReplacePlaceholders(base.Body, placeholders)

	// Do not attach CC/BCC for a test send.
	studentSettings := settings
	studentSettings.CC = nil
	studentSettings.BCC = nil

	if err := sendMailFn(studentSettings, actor.Email, subject, body); err != nil {
		return fmt.Errorf("failed to send test mail: %w", err)
	}
	return nil
}

func (s *CourseMailingService) buildSendItem(recipientID uuid.UUID, email string, base db.MailCampaign, courseInfo db.GetPassedMailingInformationRow, mailingRow db.GetParticipantMailingInformationRow, link string) campaignSendItem {
	placeholders := mailing.StatusPlaceholderValues(courseInfo.CourseName, courseInfo.CourseStartDate, courseInfo.CourseEndDate, mailingRow)
	if link != "" {
		placeholders["coursePhaseLink"] = link
	}
	return campaignSendItem{
		recipientID: recipientID,
		email:       email,
		subject:     mailing.ReplacePlaceholders(base.Subject, placeholders),
		body:        mailing.ReplacePlaceholders(base.Body, placeholders),
	}
}

// trySetCampaignSending atomically flips a campaign to "sending". A zero-row update
// is ambiguous (already sending vs. no longer existing), so on that outcome it
// re-checks existence within the same transaction to tell a genuine send-in-progress
// conflict (409) apart from a campaign deleted concurrently (404).
func (s *CourseMailingService) trySetCampaignSending(ctx context.Context, q *db.Queries, courseID, campaignID uuid.UUID) error {
	if _, err := q.TrySetMailCampaignSending(ctx, db.TrySetMailCampaignSendingParams{ID: campaignID, CourseID: courseID}); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to mark campaign as sending: %w", err)
		}
		if _, getErr := q.GetMailCampaignBase(ctx, db.GetMailCampaignBaseParams{ID: campaignID, CourseID: courseID}); errors.Is(getErr, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return ErrSendInProgress
	}
	return nil
}

func (s *CourseMailingService) recipientIDMap(ctx context.Context, campaignID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	recipients, err := s.queries.ListCampaignRecipients(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshotted recipients: %w", err)
	}
	byCP := make(map[uuid.UUID]uuid.UUID, len(recipients))
	for _, r := range recipients {
		byCP[r.CourseParticipationID] = r.ID
	}
	return byCP, nil
}

func (s *CourseMailingService) runCampaignSend(job campaignSendJob) {
	ctx, cancel := context.WithTimeout(context.Background(), campaignSendTimeout)
	defer cancel()

	// Students never get CC/BCC, so archive recipients receive exactly one copy.
	studentSettings := job.settings
	studentSettings.CC = nil
	studentSettings.BCC = nil

	sem := make(chan struct{}, maxConcurrentSends)
	var wg sync.WaitGroup
	for _, item := range job.items {
		if ctx.Err() != nil {
			log.WithField("campaignID", job.campaignID).Warn("campaign send timed out before all recipients were dispatched")
			break
		}
		item := item
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			s.sendOne(ctx, studentSettings, item)
		}()
	}
	wg.Wait()

	s.sendArchiveCopy(ctx, job)

	statusCtx := context.WithoutCancel(ctx)
	if err := s.queries.FailStalePendingRecipients(statusCtx, db.FailStalePendingRecipientsParams{
		CampaignID:   job.campaignID,
		ErrorMessage: "send did not complete",
	}); err != nil {
		log.WithError(err).WithField("campaignID", job.campaignID).Warn("failed to fail leftover pending recipients")
		return
	}

	if err := s.finalizeSend(statusCtx, job); err != nil {
		log.WithError(err).WithField("campaignID", job.campaignID).Error("failed to finalize campaign send")
	}
}

func (s *CourseMailingService) sendOne(ctx context.Context, settings mailingDTO.CourseMailingSettings, item campaignSendItem) {
	statusCtx := context.WithoutCancel(ctx)
	if item.email == "" {
		s.setRecipientFailed(statusCtx, item.recipientID, "missing email address")
		return
	}
	if err := sendMailFn(settings, item.email, item.subject, item.body); err != nil {
		// Log the recipient row ID, not the address (student email is personal data).
		log.WithError(err).WithField("recipientID", item.recipientID).Warn("failed to send campaign mail to recipient")
		s.setRecipientFailed(statusCtx, item.recipientID, err.Error())
		return
	}
	if err := s.queries.SetRecipientStatus(statusCtx, db.SetRecipientStatusParams{
		ID:           item.recipientID,
		Status:       db.MailRecipientStatusSent,
		ErrorMessage: "",
		SentAt:       pgtype.Timestamptz{Time: nowFn(), Valid: true},
	}); err != nil {
		log.WithError(err).WithField("recipientID", item.recipientID).Warn("failed to mark recipient as sent")
	}
}

func (s *CourseMailingService) setRecipientFailed(ctx context.Context, recipientID uuid.UUID, message string) {
	if err := s.queries.SetRecipientStatus(ctx, db.SetRecipientStatusParams{
		ID:           recipientID,
		Status:       db.MailRecipientStatusFailed,
		ErrorMessage: message,
		SentAt:       pgtype.Timestamptz{},
	}); err != nil {
		log.WithError(err).WithField("recipientID", recipientID).Warn("failed to mark recipient as failed")
	}
}

// sendArchiveCopy sends a single copy to the campaign's CC/BCC recipients so they
// are not copied once per student.
func (s *CourseMailingService) sendArchiveCopy(ctx context.Context, job campaignSendJob) {
	if len(job.settings.CC) == 0 && len(job.settings.BCC) == 0 {
		return
	}
	placeholders := mailing.CoursePlaceholderValues(job.courseInfo.CourseName, job.courseInfo.CourseStartDate, job.courseInfo.CourseEndDate)
	if job.link != "" {
		placeholders["coursePhaseLink"] = job.link
	}
	subject := mailing.ReplacePlaceholders(job.subject, placeholders)
	body := mailing.ReplacePlaceholders(job.body, placeholders)
	if err := sendMailFn(job.settings, job.settings.ReplyTo.Address, subject, body); err != nil {
		log.WithError(err).WithField("campaignID", job.campaignID).Warn("failed to send campaign archive copy to CC/BCC")
	}
}

func (s *CourseMailingService) finalizeSend(ctx context.Context, job campaignSendJob) error {
	recipients, err := s.queries.ListCampaignRecipients(ctx, job.campaignID)
	if err != nil {
		return fmt.Errorf("failed to load recipients for finalization: %w", err)
	}
	status := computeCampaignStatus(recipients)
	if !job.setSentMeta {
		return s.queries.SetMailCampaignStatus(ctx, db.SetMailCampaignStatusParams{ID: job.campaignID, Status: status})
	}
	return s.queries.SetMailCampaignSentMeta(ctx, db.SetMailCampaignSentMetaParams{
		ID:          job.campaignID,
		Status:      status,
		SentAt:      pgtype.Timestamptz{Time: nowFn(), Valid: true},
		SentByID:    pgtype.Text{String: job.actor.ID, Valid: true},
		SentByEmail: pgtype.Text{String: job.actor.Email, Valid: true},
		SentByName:  pgtype.Text{String: job.actor.Name, Valid: true},
	})
}

func (s *CourseMailingService) resolveSettings(ctx context.Context, base db.MailCampaign, phaseID uuid.UUID) (mailingDTO.CourseMailingSettings, error) {
	dbSettings, err := s.queries.GetCourseMailingSettingsForCoursePhaseID(ctx, phaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return mailingDTO.CourseMailingSettings{}, fmt.Errorf("%w: target course phase not found", ErrValidation)
	}
	if err != nil {
		return mailingDTO.CourseMailingSettings{}, fmt.Errorf("failed to load course mailing settings: %w", err)
	}
	settings, err := mailingDTO.GetCourseMailingSettingsFromDBModel(dbSettings)
	if err != nil {
		return mailingDTO.CourseMailingSettings{}, err
	}

	if item := parseMailItem(base.ReplyToOverride); item != nil {
		settings.ReplyTo = mail.Address{Name: item.Name, Address: item.Email}
	}
	if len(base.CcOverride) > 0 {
		settings.CC = parseAddresses(base.CcOverride)
	}
	if len(base.BccOverride) > 0 {
		settings.BCC = parseAddresses(base.BccOverride)
	}

	if settings.ReplyTo.Address == "" {
		return mailingDTO.CourseMailingSettings{}, fmt.Errorf("%w: no reply-to configured; set the course reply-to in settings or add a reply-to override", ErrValidation)
	}
	return settings, nil
}

func parseMailItem(raw []byte) *courseMailingDTO.MailItem {
	if len(raw) == 0 {
		return nil
	}
	var item courseMailingDTO.MailItem
	if err := json.Unmarshal(raw, &item); err != nil || item.Email == "" {
		return nil
	}
	return &item
}

func parseAddresses(raw []byte) []mail.Address {
	var items []courseMailingDTO.MailItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	addresses := make([]mail.Address, 0, len(items))
	for _, item := range items {
		if item.Email == "" {
			continue
		}
		addresses = append(addresses, mail.Address{Name: item.Name, Address: item.Email})
	}
	return addresses
}

// buildCoursePhaseLink resolves the management link for the {{coursePhaseLink}}
// placeholder, or returns an empty string if it cannot be resolved.
func (s *CourseMailingService) buildCoursePhaseLink(ctx context.Context, coursePhaseID uuid.UUID) string {
	courseID, err := s.queries.GetCourseIDByCoursePhaseID(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).WithField("coursePhaseID", coursePhaseID).Warn("failed to resolve course for campaign course phase link")
		return ""
	}
	return fmt.Sprintf("%s/management/course/%s/%s", s.clientURL, courseID, coursePhaseID)
}

func toMailingRow(row db.GetParticipantMailingInformationForCampaignRow) db.GetParticipantMailingInformationRow {
	return db.GetParticipantMailingInformationRow{
		FirstName:           row.FirstName,
		LastName:            row.LastName,
		Email:               row.Email,
		MatriculationNumber: row.MatriculationNumber,
		UniversityLogin:     row.UniversityLogin,
		StudyDegree:         row.StudyDegree,
		CurrentSemester:     row.CurrentSemester,
		StudyProgram:        row.StudyProgram,
	}
}
