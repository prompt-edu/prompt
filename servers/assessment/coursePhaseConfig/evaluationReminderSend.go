package coursePhaseConfig

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

var getCoreCoursePhaseFn = getCoreCoursePhase
var sendManualMailFn = sendManualMail

var (
	ErrReminderEvaluationDisabled = errors.New("evaluation type is disabled for this course phase")
	ErrReminderDeadlineNotPassed  = errors.New("evaluation deadline has not passed yet")
	ErrReminderTemplateIncomplete = errors.New("assessment reminder template is incomplete")
	errCoreRejectedMail           = errors.New("core mailing request failed")
)

const coreManualMailTimeout = 2 * time.Minute

type coreCoursePhaseResponse struct {
	ID             uuid.UUID      `json:"id"`
	Name           string         `json:"name"`
	RestrictedData map[string]any `json:"restrictedData"`
}

type coreManualMailRequest struct {
	Subject                         string            `json:"subject"`
	Content                         string            `json:"content"`
	RecipientCourseParticipationIDs []uuid.UUID       `json:"recipientCourseParticipationIDs"`
	AdditionalPlaceholders          map[string]string `json:"additionalPlaceholders"`
}

type coreManualMailReport struct {
	SuccessfulEmails    []string  `json:"successfulEmails"`
	FailedEmails        []string  `json:"failedEmails"`
	RequestedRecipients int       `json:"requestedRecipients"`
	SentAt              time.Time `json:"sentAt"`
}

func (s *CoursePhaseConfigService) SendEvaluationReminderManualTrigger(
	ctx context.Context,
	authHeader string,
	coursePhaseID uuid.UUID,
	evaluationType assessmentType.AssessmentType,
) (coursePhaseConfigDTO.EvaluationReminderSendReport, error) {
	report := coursePhaseConfigDTO.EvaluationReminderSendReport{
		SuccessfulEmails:    make([]string, 0),
		FailedEmails:        make([]string, 0),
		RequestedRecipients: 0,
		EvaluationType:      evaluationType,
	}

	recipients, err := s.getEvaluationReminderRecipients(ctx, authHeader, coursePhaseID, evaluationType)
	if err != nil {
		return report, err
	}
	report.Deadline = recipients.Deadline
	report.DeadlinePassed = recipients.DeadlinePassed

	if !recipients.EvaluationEnabled {
		return report, ErrReminderEvaluationDisabled
	}
	if !recipients.DeadlinePassed {
		if recipients.Deadline != nil {
			return report, fmt.Errorf("%w (deadline: %s)", ErrReminderDeadlineNotPassed, recipients.Deadline.Format(time.RFC3339))
		}
		return report, fmt.Errorf("%w: deadline is not configured", ErrReminderDeadlineNotPassed)
	}

	coursePhase, err := getCoreCoursePhaseFn(ctx, authHeader, coursePhaseID)
	if err != nil {
		return report, err
	}

	subject, content, legacyLastSentByType := getAssessmentReminderTemplate(coursePhase.RestrictedData)
	if subject == "" || content == "" {
		return report, ErrReminderTemplateIncomplete
	}
	report.PreviousSentAt, err = s.getPreviousReminderSentAt(ctx, coursePhaseID, evaluationType, legacyLastSentByType)
	if err != nil {
		return report, err
	}

	mailReport, err := sendManualMailFn(ctx, authHeader, coursePhaseID, coreManualMailRequest{
		Subject:                         subject,
		Content:                         content,
		RecipientCourseParticipationIDs: recipients.IncompleteAuthorCourseParticipationIDs,
		AdditionalPlaceholders: map[string]string{
			"evaluationType":     recipients.EvaluationTypeLabel,
			"evaluationDeadline": recipients.EvaluationDeadlinePlaceholder,
			"coursePhaseName":    coursePhase.Name,
		},
	})
	if err != nil {
		return report, err
	}

	report.SuccessfulEmails = mailReport.SuccessfulEmails
	report.FailedEmails = mailReport.FailedEmails
	report.RequestedRecipients = mailReport.RequestedRecipients
	report.SentAt = mailReport.SentAt

	// The reminder state lives in this service's database, so persisting it cannot overwrite the
	// course phase settings in core that were changed while the mails were being sent.
	if err := s.queries.UpsertEvaluationReminderLastSentAt(ctx, db.UpsertEvaluationReminderLastSentAtParams{
		CoursePhaseID:  coursePhaseID,
		EvaluationType: assessmentType.MapDTOtoDBAssessmentType(evaluationType),
		LastSentAt:     pgtype.Timestamptz{Time: report.SentAt, Valid: true},
	}); err != nil {
		log.WithError(err).
			WithField("coursePhaseID", coursePhaseID).
			WithField("evaluationType", evaluationType).
			Warn("Evaluation reminder mails were sent, but persisting lastSentAt failed")
		return report, nil
	}

	return report, nil
}

// GetEvaluationReminderStatus returns when a reminder was last sent for each evaluation type.
func (s *CoursePhaseConfigService) GetEvaluationReminderStatus(ctx context.Context, coursePhaseID uuid.UUID) (coursePhaseConfigDTO.EvaluationReminderStatus, error) {
	reminders, err := s.queries.GetEvaluationRemindersForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseConfigDTO.EvaluationReminderStatus{}, fmt.Errorf("failed to get evaluation reminders: %w", err)
	}

	status := coursePhaseConfigDTO.EvaluationReminderStatus{
		LastSentAtByType: make(map[assessmentType.AssessmentType]time.Time, len(reminders)),
	}
	for _, reminder := range reminders {
		status.LastSentAtByType[assessmentType.MapDBAssessmentTypeToDTO(reminder.EvaluationType)] = reminder.LastSentAt.Time
	}
	return status, nil
}

// getPreviousReminderSentAt prefers the stored reminder state and falls back to the
// lastSentAtByType that older versions kept in the course phase's restricted data.
func (s *CoursePhaseConfigService) getPreviousReminderSentAt(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	evaluationType assessmentType.AssessmentType,
	legacyLastSentByType map[string]string,
) (*time.Time, error) {
	reminder, err := s.queries.GetEvaluationReminder(ctx, db.GetEvaluationReminderParams{
		CoursePhaseID:  coursePhaseID,
		EvaluationType: assessmentType.MapDTOtoDBAssessmentType(evaluationType),
	})
	if err == nil {
		return &reminder.LastSentAt.Time, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get evaluation reminder: %w", err)
	}
	return getLegacyReminderSentAt(legacyLastSentByType, evaluationType), nil
}

func getCoreCoursePhase(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (coreCoursePhaseResponse, error) {
	endpoint := fmt.Sprintf("%s/api/course_phases/%s", sdkUtils.GetCoreUrl(), coursePhaseID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return coreCoursePhaseResponse{}, fmt.Errorf("failed to create core course phase request: %w", err)
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return coreCoursePhaseResponse{}, fmt.Errorf("failed to fetch course phase from core: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return coreCoursePhaseResponse{}, fmt.Errorf("failed to read course phase response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return coreCoursePhaseResponse{}, fmt.Errorf(
			"core course phase request failed with status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var parsed coreCoursePhaseResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return coreCoursePhaseResponse{}, fmt.Errorf("failed to parse core course phase response: %w", err)
	}
	if parsed.RestrictedData == nil {
		parsed.RestrictedData = map[string]any{}
	}

	return parsed, nil
}

func sendManualMail(
	ctx context.Context,
	authHeader string,
	coursePhaseID uuid.UUID,
	request coreManualMailRequest,
) (coreManualMailReport, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return coreManualMailReport{}, fmt.Errorf("failed to marshal manual mail request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/mailing/%s/manual", sdkUtils.GetCoreUrl(), coursePhaseID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return coreManualMailReport{}, fmt.Errorf("failed to create core mailing request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{Timeout: coreManualMailTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return coreManualMailReport{}, fmt.Errorf("failed to send manual mails via core: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return coreManualMailReport{}, fmt.Errorf(
			"%w with status %d: %s",
			errCoreRejectedMail,
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}
	if readErr != nil {
		return coreManualMailReport{}, fmt.Errorf("failed to read core mailing response: %w", readErr)
	}

	var parsed coreManualMailReport
	if err := json.Unmarshal(body, &parsed); err != nil {
		return coreManualMailReport{}, fmt.Errorf("failed to parse core mailing response: %w", err)
	}
	return parsed, nil
}

func getAssessmentReminderTemplate(restrictedData map[string]any) (string, string, map[string]string) {
	mailingSettings, ok := restrictedData["mailingSettings"].(map[string]any)
	if !ok {
		return "", "", map[string]string{}
	}
	assessmentReminder, ok := mailingSettings["assessmentReminder"].(map[string]any)
	if !ok {
		return "", "", map[string]string{}
	}

	subject, _ := assessmentReminder["subject"].(string)
	content, _ := assessmentReminder["content"].(string)

	lastSentByType := make(map[string]string)
	rawLastSentByType, ok := assessmentReminder["lastSentAtByType"].(map[string]any)
	if ok {
		for key, value := range rawLastSentByType {
			if parsed, valueOk := value.(string); valueOk {
				lastSentByType[key] = parsed
			}
		}
	}

	return subject, content, lastSentByType
}

func getLegacyReminderSentAt(lastSentByType map[string]string, evaluationType assessmentType.AssessmentType) *time.Time {
	if len(lastSentByType) == 0 {
		return nil
	}
	rawValue, ok := lastSentByType[string(evaluationType)]
	if !ok || rawValue == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, rawValue)
	if err != nil {
		return nil
	}
	return &parsed
}
