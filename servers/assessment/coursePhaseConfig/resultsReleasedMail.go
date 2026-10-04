package coursePhaseConfig

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

func (s *CoursePhaseConfigService) SendResultsReleasedMail(
	ctx context.Context,
	authHeader string,
	coursePhaseID uuid.UUID,
) (*coursePhaseConfigDTO.ResultsReleasedMailReport, error) {
	coursePhase, err := getCoreCoursePhaseFn(ctx, authHeader, coursePhaseID)
	if err != nil {
		return nil, err
	}

	subject, content, sendOnRelease := getResultsReleasedMailTemplate(coursePhase.RestrictedData)
	if !sendOnRelease || subject == "" || content == "" {
		return nil, nil
	}

	recipients, err := s.getResultsReleasedMailRecipients(ctx, authHeader, coursePhaseID)
	if err != nil {
		return nil, err
	}

	claimedIDs, err := s.queries.ClaimResultsReleasedMailRecipients(ctx, db.ClaimResultsReleasedMailRecipientsParams{
		CoursePhaseID:          coursePhaseID,
		CourseParticipationIds: getCourseParticipationIDs(recipients),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim results mail recipients: %w", err)
	}

	report := &coursePhaseConfigDTO.ResultsReleasedMailReport{
		SuccessfulEmails: make([]string, 0),
		FailedEmails:     make([]string, 0),
	}
	if len(claimedIDs) == 0 {
		return report, nil
	}

	mailReport, err := sendManualMailFn(ctx, authHeader, coursePhaseID, coreManualMailRequest{
		Subject:                         subject,
		Content:                         content,
		RecipientCourseParticipationIDs: claimedIDs,
		AdditionalPlaceholders:          map[string]string{"coursePhaseName": coursePhase.Name},
	})
	if err != nil {
		s.releaseResultsReleasedMailClaims(ctx, coursePhaseID, claimedIDs)
		return nil, err
	}
	s.releaseResultsReleasedMailClaims(ctx, coursePhaseID, getCourseParticipationIDsByEmail(recipients, mailReport.FailedEmails))

	report.SuccessfulEmails = mailReport.SuccessfulEmails
	report.FailedEmails = mailReport.FailedEmails
	report.RequestedRecipients = mailReport.RequestedRecipients
	return report, nil
}

func (s *CoursePhaseConfigService) getResultsReleasedMailRecipients(
	ctx context.Context,
	authHeader string,
	coursePhaseID uuid.UUID,
) ([]coursePhaseConfigDTO.AssessmentParticipationWithStudent, error) {
	config, err := s.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to load course phase config: %w", err)
	}

	participations, err := getParticipationsForCoursePhaseFn(ctx, authHeader, coursePhaseID)
	if err != nil {
		return nil, err
	}
	if !config.AssessmentEnabled {
		return participations, nil
	}

	finalAssessments, err := s.queries.GetAllGrades(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to load final assessments: %w", err)
	}
	isFinal := make(map[uuid.UUID]bool, len(finalAssessments))
	for _, assessment := range finalAssessments {
		isFinal[assessment.CourseParticipationID] = true
	}

	return slices.DeleteFunc(participations, func(participation coursePhaseConfigDTO.AssessmentParticipationWithStudent) bool {
		return !isFinal[participation.CourseParticipationID]
	}), nil
}

func (s *CoursePhaseConfigService) releaseResultsReleasedMailClaims(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationIDs []uuid.UUID) {
	if len(courseParticipationIDs) == 0 {
		return
	}
	err := s.queries.ReleaseResultsReleasedMailClaims(ctx, db.ReleaseResultsReleasedMailClaimsParams{
		CoursePhaseID:          coursePhaseID,
		CourseParticipationIds: courseParticipationIDs,
	})
	if err != nil {
		log.WithError(err).
			WithField("coursePhaseID", coursePhaseID).
			Warn("Results mail was not delivered, but releasing the recipient claims failed")
	}
}

func getResultsReleasedMailTemplate(restrictedData map[string]any) (subject string, content string, sendOnRelease bool) {
	mailingSettings, _ := restrictedData["mailingSettings"].(map[string]any)
	resultsReleasedMail, _ := mailingSettings["resultsReleasedMail"].(map[string]any)

	subject, _ = resultsReleasedMail["subject"].(string)
	content, _ = resultsReleasedMail["content"].(string)
	sendOnRelease, _ = resultsReleasedMail["sendOnRelease"].(bool)
	return subject, content, sendOnRelease
}

func getCourseParticipationIDs(participations []coursePhaseConfigDTO.AssessmentParticipationWithStudent) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(participations))
	for _, participation := range participations {
		ids = append(ids, participation.CourseParticipationID)
	}
	return ids
}

func getCourseParticipationIDsByEmail(
	participations []coursePhaseConfigDTO.AssessmentParticipationWithStudent,
	emails []string,
) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(emails))
	for _, participation := range participations {
		if slices.Contains(emails, participation.Student.Email) {
			ids = append(ids, participation.CourseParticipationID)
		}
	}
	return ids
}
