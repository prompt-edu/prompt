package mailing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
)

var sendMailFn = SendMail
var nowFn = func() time.Time {
	return time.Now().UTC()
}

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
		return report, fmt.Errorf("mailing template incomplete: subject or content is empty")
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
		placeholderMap := getStatusEmailPlaceholderValues(
			passedMailingInfo.CourseName,
			passedMailingInfo.CourseStartDate,
			passedMailingInfo.CourseEndDate,
			db.GetParticipantMailingInformationRow{
				FirstName:           participant.FirstName,
				LastName:            participant.LastName,
				Email:               participant.Email,
				MatriculationNumber: participant.MatriculationNumber,
				UniversityLogin:     participant.UniversityLogin,
				StudyDegree:         participant.StudyDegree,
				CurrentSemester:     participant.CurrentSemester,
				StudyProgram:        participant.StudyProgram,
			},
		)
		for key, value := range request.AdditionalPlaceholders {
			placeholderMap[key] = value
		}

		finalSubject := replacePlaceholders(request.Subject, placeholderMap)
		finalContent := replacePlaceholders(request.Content, placeholderMap)

		if err := sendMailFn(courseMailingSettings, participant.Email.String, finalSubject, finalContent); err != nil {
			report.FailedEmails = append(report.FailedEmails, participant.Email.String)
			continue
		}
		report.SuccessfulEmails = append(report.SuccessfulEmails, participant.Email.String)
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
