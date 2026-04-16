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
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	log "github.com/sirupsen/logrus"
)

var getEvaluationReminderRecipientsForSendFn = GetEvaluationReminderRecipients
var getCoreCoursePhaseFn = getCoreCoursePhase
var sendManualReminderMailFn = sendManualReminderMail
var updateCoreCoursePhaseFn = updateCoreCoursePhase

var (
	ErrReminderEvaluationDisabled = errors.New("evaluation type is disabled for this course phase")
	ErrReminderDeadlineNotPassed  = errors.New("evaluation deadline has not passed yet")
)

const coreManualMailTimeout = 2 * time.Minute

type coreCoursePhaseResponse struct {
	ID                  uuid.UUID      `json:"id"`
	Name                string         `json:"name"`
	RestrictedData      map[string]any `json:"restrictedData"`
	StudentReadableData map[string]any `json:"studentReadableData"`
}

type coreUpdateCoursePhaseRequest struct {
	ID                  uuid.UUID      `json:"id"`
	Name                string         `json:"name"`
	RestrictedData      map[string]any `json:"restrictedData"`
	StudentReadableData map[string]any `json:"studentReadableData"`
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

func SendEvaluationReminderManualTrigger(
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

	recipients, err := getEvaluationReminderRecipientsForSendFn(ctx, authHeader, coursePhaseID, evaluationType)
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
		return report, ErrReminderDeadlineNotPassed
	}

	coursePhase, err := getCoreCoursePhaseFn(ctx, authHeader, coursePhaseID)
	if err != nil {
		return report, err
	}

	subject, content, lastSentByType := getAssessmentReminderTemplate(coursePhase.RestrictedData)
	if subject == "" || content == "" {
		return report, fmt.Errorf("assessment reminder template incomplete: subject or content is empty")
	}
	report.PreviousSentAt = getPreviousReminderSentAt(lastSentByType, evaluationType)

	mailReport, err := sendManualReminderMailFn(ctx, authHeader, coursePhaseID, coreManualMailRequest{
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

	setLastSentAt(coursePhase.RestrictedData, evaluationType, report.SentAt)
	if err := updateCoreCoursePhaseFn(ctx, authHeader, coreUpdateCoursePhaseRequest(coursePhase)); err != nil {
		log.WithError(err).
			WithField("coursePhaseID", coursePhase.ID).
			WithField("evaluationType", evaluationType).
			Warn("Evaluation reminder mails were sent, but persisting lastSentAt failed")
		return report, nil
	}

	return report, nil
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
	if parsed.StudentReadableData == nil {
		parsed.StudentReadableData = map[string]any{}
	}

	return parsed, nil
}

func sendManualReminderMail(
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return coreManualMailReport{}, fmt.Errorf("failed to read core mailing response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return coreManualMailReport{}, fmt.Errorf(
			"core mailing request failed with status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var parsed coreManualMailReport
	if err := json.Unmarshal(body, &parsed); err != nil {
		return coreManualMailReport{}, fmt.Errorf("failed to parse core mailing response: %w", err)
	}
	return parsed, nil
}

func updateCoreCoursePhase(
	ctx context.Context,
	authHeader string,
	request coreUpdateCoursePhaseRequest,
) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal course phase update request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/course_phases/%s", sdkUtils.GetCoreUrl(), request.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create core course phase update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update course phase in core: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read core course phase update response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"core course phase update failed with status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	return nil
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

func getPreviousReminderSentAt(lastSentByType map[string]string, evaluationType assessmentType.AssessmentType) *time.Time {
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

func setLastSentAt(restrictedData map[string]any, evaluationType assessmentType.AssessmentType, sentAt time.Time) {
	mailingSettings := getOrCreateMap(restrictedData, "mailingSettings")
	assessmentReminder := getOrCreateMap(mailingSettings, "assessmentReminder")
	lastSentAtByType := getOrCreateMap(assessmentReminder, "lastSentAtByType")
	lastSentAtByType[string(evaluationType)] = sentAt.Format(time.RFC3339)
}

func getOrCreateMap(parent map[string]any, key string) map[string]any {
	existing, ok := parent[key].(map[string]any)
	if ok {
		return existing
	}
	created := map[string]any{}
	parent[key] = created
	return created
}
