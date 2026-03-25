package mailing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	log "github.com/sirupsen/logrus"
)

var sendMailFn = SendMail
var nowFn = func() time.Time {
	return time.Now().UTC()
}

var ErrManualMailValidation = errors.New("manual mail validation error")

func SendManualMailToParticipants(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	request mailingDTO.SendManualMailRequest,
) (mailingDTO.ManualMailReport, error) {
	report := mailingDTO.ManualMailReport{
		SuccessfulEmails:    make([]string, 0),
		FailedEmails:        make([]string, 0),
		RequestedRecipients: 0,
	}

	if request.Subject == "" || request.Content == "" {
		return report, fmt.Errorf("%w: mailing template incomplete: subject or content is empty", ErrManualMailValidation)
	}

	recipientIDs := deduplicateUUIDList(request.RecipientCourseParticipationIDs)
	participants, err := MailingServiceSingleton.queries.GetParticipantMailingInformationByIDs(
		ctx,
		db.GetParticipantMailingInformationByIDsParams{
			ID:      coursePhaseID,
			Column2: recipientIDs,
		},
	)
	if err != nil {
		return report, fmt.Errorf("failed to load mail recipients: %w", err)
	}
	report.RequestedRecipients = len(participants)

	courseMailingSettings, err := getSenderInformation(ctx, coursePhaseID)
	if err != nil {
		return report, fmt.Errorf("failed to get sender information: %w", err)
	}

	passedMailingInfo, err := MailingServiceSingleton.queries.GetPassedMailingInformation(ctx, coursePhaseID)
	if err != nil {
		return report, fmt.Errorf("failed to load course mailing context: %w", err)
	}

	for _, participant := range participants {
		if !participant.Email.Valid || participant.Email.String == "" {
			failedIdentifier := participant.UniversityLogin.String
			if failedIdentifier == "" {
				failedIdentifier = participant.MatriculationNumber.String
			}
			if failedIdentifier == "" {
				failedIdentifier = "missing-email"
			}
			log.WithField("coursePhaseID", coursePhaseID).
				WithField("participantIdentifier", failedIdentifier).
				Warn("Skipping manual mail recipient with missing email address")
			report.FailedEmails = append(report.FailedEmails, failedIdentifier)
			continue
		}

		placeholderMap := getStatusEmailPlaceholderValues(
			passedMailingInfo.CourseName,
			passedMailingInfo.CourseStartDate,
			passedMailingInfo.CourseEndDate,
			db.GetParticipantMailingInformationRow(participant),
		)
		for key, value := range request.AdditionalPlaceholders {
			placeholderMap[key] = value
		}

		finalSubject := replacePlaceholders(request.Subject, placeholderMap)
		finalContent := replacePlaceholders(request.Content, placeholderMap)

		recipientEmail := participant.Email.String
		if err := sendMailFn(courseMailingSettings, recipientEmail, finalSubject, finalContent); err != nil {
			log.WithError(err).WithField("email", recipientEmail).Warn("Failed to send manual mail to participant")
			report.FailedEmails = append(report.FailedEmails, recipientEmail)
			continue
		}
		report.SuccessfulEmails = append(report.SuccessfulEmails, recipientEmail)
	}

	report.SentAt = nowFn()
	return report, nil
}

func deduplicateUUIDList(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
