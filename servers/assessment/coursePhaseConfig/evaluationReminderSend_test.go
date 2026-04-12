package coursePhaseConfig

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendEvaluationReminderManualTriggerHappyPath(t *testing.T) {
	oldRecipientsFn := getEvaluationReminderRecipientsForSendFn
	oldGetCoreCoursePhaseFn := getCoreCoursePhaseFn
	oldSendManualReminderMailFn := sendManualReminderMailFn
	oldUpdateCoreCoursePhaseFn := updateCoreCoursePhaseFn
	t.Cleanup(func() {
		getEvaluationReminderRecipientsForSendFn = oldRecipientsFn
		getCoreCoursePhaseFn = oldGetCoreCoursePhaseFn
		sendManualReminderMailFn = oldSendManualReminderMailFn
		updateCoreCoursePhaseFn = oldUpdateCoreCoursePhaseFn
	})

	ctx := context.Background()
	coursePhaseID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	recipientID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	deadline := time.Date(2026, time.January, 9, 15, 0, 0, 0, time.UTC)
	previousSentAt := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)
	sentAt := time.Date(2026, time.January, 10, 10, 0, 0, 0, time.UTC)

	getEvaluationReminderRecipientsForSendFn = func(
		ctx context.Context,
		authHeader string,
		coursePhaseID uuid.UUID,
		evaluationType assessmentType.AssessmentType,
	) (coursePhaseConfigDTO.EvaluationReminderRecipients, error) {
		return coursePhaseConfigDTO.EvaluationReminderRecipients{
			EvaluationType:                         evaluationType,
			EvaluationTypeLabel:                    "Self Evaluation",
			EvaluationEnabled:                      true,
			Deadline:                               &deadline,
			EvaluationDeadlinePlaceholder:          "09.01.2026 15:00",
			DeadlinePassed:                         true,
			IncompleteAuthorCourseParticipationIDs: []uuid.UUID{recipientID},
		}, nil
	}

	getCoreCoursePhaseFn = func(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (coreCoursePhaseResponse, error) {
		return coreCoursePhaseResponse{
			ID:   coursePhaseID,
			Name: "Assessment Phase",
			RestrictedData: map[string]any{
				"mailingSettings": map[string]any{
					"assessmentReminder": map[string]any{
						"subject": "Reminder {{evaluationType}}",
						"content": "Please finish {{evaluationType}} in {{coursePhaseName}} before {{evaluationDeadline}}",
						"lastSentAtByType": map[string]any{
							"self": previousSentAt.Format(time.RFC3339),
						},
					},
				},
			},
			StudentReadableData: map[string]any{},
		}, nil
	}

	var capturedMailRequest coreManualMailRequest
	sendManualReminderMailFn = func(
		ctx context.Context,
		authHeader string,
		coursePhaseID uuid.UUID,
		request coreManualMailRequest,
	) (coreManualMailReport, error) {
		capturedMailRequest = request
		return coreManualMailReport{
			SuccessfulEmails:    []string{"alice@example.com"},
			FailedEmails:        []string{},
			RequestedRecipients: 1,
			SentAt:              sentAt,
		}, nil
	}

	var capturedUpdateRequest coreUpdateCoursePhaseRequest
	updateCoreCoursePhaseFn = func(ctx context.Context, authHeader string, request coreUpdateCoursePhaseRequest) error {
		capturedUpdateRequest = request
		return nil
	}

	report, err := SendEvaluationReminderManualTrigger(ctx, "Bearer token", coursePhaseID, assessmentType.Self)
	require.NoError(t, err)
	require.NotNil(t, report.PreviousSentAt)
	assert.Equal(t, previousSentAt.Format(time.RFC3339), report.PreviousSentAt.UTC().Format(time.RFC3339))
	assert.Equal(t, sentAt, report.SentAt)
	assert.Equal(t, 1, report.RequestedRecipients)
	assert.Equal(t, []uuid.UUID{recipientID}, capturedMailRequest.RecipientCourseParticipationIDs)
	assert.Equal(t, "Self Evaluation", capturedMailRequest.AdditionalPlaceholders["evaluationType"])
	assert.Equal(t, "Assessment Phase", capturedMailRequest.AdditionalPlaceholders["coursePhaseName"])
	assert.Equal(t, "09.01.2026 15:00", capturedMailRequest.AdditionalPlaceholders["evaluationDeadline"])

	mailingSettings := capturedUpdateRequest.RestrictedData["mailingSettings"].(map[string]any)
	assessmentReminder := mailingSettings["assessmentReminder"].(map[string]any)
	lastSentAtByType := assessmentReminder["lastSentAtByType"].(map[string]any)
	assert.Equal(t, sentAt.Format(time.RFC3339), lastSentAtByType["self"])
}

func TestSendEvaluationReminderManualTriggerDeadlineNotPassed(t *testing.T) {
	oldRecipientsFn := getEvaluationReminderRecipientsForSendFn
	t.Cleanup(func() {
		getEvaluationReminderRecipientsForSendFn = oldRecipientsFn
	})

	deadline := time.Date(2026, time.January, 20, 15, 0, 0, 0, time.UTC)
	getEvaluationReminderRecipientsForSendFn = func(
		ctx context.Context,
		authHeader string,
		coursePhaseID uuid.UUID,
		evaluationType assessmentType.AssessmentType,
	) (coursePhaseConfigDTO.EvaluationReminderRecipients, error) {
		return coursePhaseConfigDTO.EvaluationReminderRecipients{
			EvaluationEnabled: true,
			Deadline:          &deadline,
			DeadlinePassed:    false,
		}, nil
	}

	_, err := SendEvaluationReminderManualTrigger(context.Background(), "Bearer token", uuid.New(), assessmentType.Self)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deadline")
}

func TestSendEvaluationReminderManualTriggerUpdateFailureStillSucceeds(t *testing.T) {
	oldRecipientsFn := getEvaluationReminderRecipientsForSendFn
	oldGetCoreCoursePhaseFn := getCoreCoursePhaseFn
	oldSendManualReminderMailFn := sendManualReminderMailFn
	oldUpdateCoreCoursePhaseFn := updateCoreCoursePhaseFn
	t.Cleanup(func() {
		getEvaluationReminderRecipientsForSendFn = oldRecipientsFn
		getCoreCoursePhaseFn = oldGetCoreCoursePhaseFn
		sendManualReminderMailFn = oldSendManualReminderMailFn
		updateCoreCoursePhaseFn = oldUpdateCoreCoursePhaseFn
	})

	ctx := context.Background()
	coursePhaseID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	recipientID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	deadline := time.Date(2026, time.January, 9, 15, 0, 0, 0, time.UTC)
	sentAt := time.Date(2026, time.January, 10, 10, 0, 0, 0, time.UTC)

	getEvaluationReminderRecipientsForSendFn = func(
		ctx context.Context,
		authHeader string,
		coursePhaseID uuid.UUID,
		evaluationType assessmentType.AssessmentType,
	) (coursePhaseConfigDTO.EvaluationReminderRecipients, error) {
		return coursePhaseConfigDTO.EvaluationReminderRecipients{
			EvaluationType:                         evaluationType,
			EvaluationTypeLabel:                    "Self Evaluation",
			EvaluationEnabled:                      true,
			Deadline:                               &deadline,
			EvaluationDeadlinePlaceholder:          "09.01.2026 15:00",
			DeadlinePassed:                         true,
			IncompleteAuthorCourseParticipationIDs: []uuid.UUID{recipientID},
		}, nil
	}

	getCoreCoursePhaseFn = func(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (coreCoursePhaseResponse, error) {
		return coreCoursePhaseResponse{
			ID:   coursePhaseID,
			Name: "Assessment Phase",
			RestrictedData: map[string]any{
				"mailingSettings": map[string]any{
					"assessmentReminder": map[string]any{
						"subject": "Reminder {{evaluationType}}",
						"content": "Please finish {{evaluationType}} in {{coursePhaseName}} before {{evaluationDeadline}}",
					},
				},
			},
			StudentReadableData: map[string]any{},
		}, nil
	}

	sendManualReminderMailFn = func(
		ctx context.Context,
		authHeader string,
		coursePhaseID uuid.UUID,
		request coreManualMailRequest,
	) (coreManualMailReport, error) {
		return coreManualMailReport{
			SuccessfulEmails:    []string{"alice@example.com"},
			FailedEmails:        []string{},
			RequestedRecipients: 1,
			SentAt:              sentAt,
		}, nil
	}

	updateCoreCoursePhaseFn = func(ctx context.Context, authHeader string, request coreUpdateCoursePhaseRequest) error {
		return errors.New("failed to persist")
	}

	report, err := SendEvaluationReminderManualTrigger(ctx, "Bearer token", coursePhaseID, assessmentType.Self)
	require.NoError(t, err)
	assert.Equal(t, sentAt, report.SentAt)
	assert.Equal(t, 1, report.RequestedRecipients)
	assert.Equal(t, []string{"alice@example.com"}, report.SuccessfulEmails)
}
