package coursePhaseConfig

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var reminderRecipientID = uuid.MustParse("22222222-2222-2222-2222-222222222222")

type EvaluationReminderSendTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	service *CoursePhaseConfigService

	restoreStubs   func()
	restrictedData map[string]any
	mailRequests   []coreManualMailRequest
	sentAt         time.Time
}

func (suite *EvaluationReminderSendTestSuite) SetupSuite() {
	if testing.Short() {
		suite.T().Skip("skipping db-backed evaluation reminder tests in short mode")
	}
	defer func() {
		if r := recover(); r != nil {
			suite.T().Skipf("skipping db-backed evaluation reminder tests: %v", r)
		}
	}()

	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/coursePhaseConfig.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Skipf("skipping db-backed evaluation reminder tests: %v", err)
	}
	suite.cleanup = cleanup
	suite.service = NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn, nil)
	suite.service.getEvaluationReminderRecipients = func(
		ctx context.Context,
		authHeader string,
		coursePhaseID uuid.UUID,
		evaluationType assessmentType.AssessmentType,
	) (coursePhaseConfigDTO.EvaluationReminderRecipients, error) {
		deadline := time.Date(2026, time.January, 9, 15, 0, 0, 0, time.UTC)
		return coursePhaseConfigDTO.EvaluationReminderRecipients{
			EvaluationType:                         evaluationType,
			EvaluationTypeLabel:                    "Self Evaluation",
			EvaluationEnabled:                      true,
			Deadline:                               &deadline,
			EvaluationDeadlinePlaceholder:          "09.01.2026 15:00",
			DeadlinePassed:                         true,
			IncompleteAuthorCourseParticipationIDs: []uuid.UUID{reminderRecipientID},
		}, nil
	}
}

func (suite *EvaluationReminderSendTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *EvaluationReminderSendTestSuite) SetupTest() {
	suite.restrictedData = reminderRestrictedData(nil)
	suite.mailRequests = nil
	suite.sentAt = time.Date(2026, time.January, 10, 10, 0, 0, 0, time.UTC)

	oldGetCoreCoursePhaseFn := getCoreCoursePhaseFn
	oldSendManualMailFn := sendManualMailFn
	suite.restoreStubs = func() {
		getCoreCoursePhaseFn = oldGetCoreCoursePhaseFn
		sendManualMailFn = oldSendManualMailFn
	}

	getCoreCoursePhaseFn = func(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (coreCoursePhaseResponse, error) {
		return coreCoursePhaseResponse{
			ID:             coursePhaseID,
			Name:           "Assessment Phase",
			RestrictedData: suite.restrictedData,
		}, nil
	}
	sendManualMailFn = func(ctx context.Context, authHeader string, coursePhaseID uuid.UUID, request coreManualMailRequest) (coreManualMailReport, error) {
		suite.mailRequests = append(suite.mailRequests, request)
		return coreManualMailReport{
			SuccessfulEmails:    []string{"alice@example.com"},
			FailedEmails:        []string{},
			RequestedRecipients: len(request.RecipientCourseParticipationIDs),
			SentAt:              suite.sentAt,
		}, nil
	}
}

func (suite *EvaluationReminderSendTestSuite) TearDownTest() {
	suite.restoreStubs()
}

func reminderRestrictedData(lastSentAtByType map[string]any) map[string]any {
	assessmentReminder := map[string]any{
		"subject": "Reminder {{evaluationType}}",
		"content": "Please finish {{evaluationType}} in {{coursePhaseName}} before {{evaluationDeadline}}",
	}
	if lastSentAtByType != nil {
		assessmentReminder["lastSentAtByType"] = lastSentAtByType
	}
	return map[string]any{
		"mailingSettings": map[string]any{"assessmentReminder": assessmentReminder},
	}
}

func (suite *EvaluationReminderSendTestSuite) lastSentAt(coursePhaseID uuid.UUID, evaluationType assessmentType.AssessmentType) (time.Time, bool) {
	status, err := suite.service.GetEvaluationReminderStatus(suite.ctx, coursePhaseID)
	suite.Require().NoError(err)
	sentAt, ok := status.LastSentAtByType[evaluationType]
	return sentAt, ok
}

func (suite *EvaluationReminderSendTestSuite) send(coursePhaseID uuid.UUID) coursePhaseConfigDTO.EvaluationReminderSendReport {
	report, err := suite.service.SendEvaluationReminderManualTrigger(suite.ctx, "Bearer token", coursePhaseID, assessmentType.Self)
	suite.Require().NoError(err)
	return report
}

func (suite *EvaluationReminderSendTestSuite) TestSendsMailAndStoresLastSentAt() {
	coursePhaseID := uuid.New()

	report := suite.send(coursePhaseID)
	suite.Nil(report.PreviousSentAt)
	suite.Equal(suite.sentAt, report.SentAt)
	suite.Equal(1, report.RequestedRecipients)

	suite.Require().Len(suite.mailRequests, 1)
	mail := suite.mailRequests[0]
	suite.Equal([]uuid.UUID{reminderRecipientID}, mail.RecipientCourseParticipationIDs)
	suite.Equal("Self Evaluation", mail.AdditionalPlaceholders["evaluationType"])
	suite.Equal("Assessment Phase", mail.AdditionalPlaceholders["coursePhaseName"])
	suite.Equal("09.01.2026 15:00", mail.AdditionalPlaceholders["evaluationDeadline"])

	sentAt, ok := suite.lastSentAt(coursePhaseID, assessmentType.Self)
	suite.Require().True(ok)
	suite.True(suite.sentAt.Equal(sentAt))
	_, ok = suite.lastSentAt(coursePhaseID, assessmentType.Peer)
	suite.False(ok)
}

func (suite *EvaluationReminderSendTestSuite) TestRepeatedSendReportsAndReplacesPreviousSentAt() {
	coursePhaseID := uuid.New()
	firstSentAt := suite.sentAt
	suite.send(coursePhaseID)

	suite.sentAt = firstSentAt.Add(24 * time.Hour)
	report := suite.send(coursePhaseID)
	suite.Require().NotNil(report.PreviousSentAt)
	suite.True(firstSentAt.Equal(*report.PreviousSentAt))

	sentAt, ok := suite.lastSentAt(coursePhaseID, assessmentType.Self)
	suite.Require().True(ok)
	suite.True(suite.sentAt.Equal(sentAt))
}

func (suite *EvaluationReminderSendTestSuite) TestFallsBackToLegacyLastSentAt() {
	coursePhaseID := uuid.New()
	legacySentAt := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)
	suite.restrictedData = reminderRestrictedData(map[string]any{"self": legacySentAt.Format(time.RFC3339)})

	report := suite.send(coursePhaseID)
	suite.Require().NotNil(report.PreviousSentAt)
	suite.True(legacySentAt.Equal(*report.PreviousSentAt))

	// Once a send is stored, it wins over the legacy value.
	firstSentAt := suite.sentAt
	suite.sentAt = firstSentAt.Add(time.Hour)
	report = suite.send(coursePhaseID)
	suite.Require().NotNil(report.PreviousSentAt)
	suite.True(firstSentAt.Equal(*report.PreviousSentAt))
}

func (suite *EvaluationReminderSendTestSuite) TestOlderSendDoesNotReplaceNewerLastSentAt() {
	coursePhaseID := uuid.New()
	newerSentAt := suite.sentAt
	suite.send(coursePhaseID)

	suite.sentAt = newerSentAt.Add(-time.Hour)
	suite.send(coursePhaseID)

	sentAt, ok := suite.lastSentAt(coursePhaseID, assessmentType.Self)
	suite.Require().True(ok)
	suite.True(newerSentAt.Equal(sentAt))
}

func (suite *EvaluationReminderSendTestSuite) TestPersistFailureStillReturnsReport() {
	_, err := suite.service.conn.Exec(suite.ctx, `
		CREATE FUNCTION reject_evaluation_reminder() RETURNS trigger AS $$
		BEGIN RAISE EXCEPTION 'evaluation_reminder unavailable'; END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER reject_evaluation_reminder BEFORE INSERT OR UPDATE ON evaluation_reminder
			FOR EACH ROW EXECUTE FUNCTION reject_evaluation_reminder();`)
	suite.Require().NoError(err)
	suite.T().Cleanup(func() {
		_, err := suite.service.conn.Exec(suite.ctx, `
			DROP TRIGGER reject_evaluation_reminder ON evaluation_reminder;
			DROP FUNCTION reject_evaluation_reminder();`)
		suite.Require().NoError(err)
	})

	coursePhaseID := uuid.New()
	report := suite.send(coursePhaseID)
	suite.Equal(suite.sentAt, report.SentAt)
	suite.Equal([]string{"alice@example.com"}, report.SuccessfulEmails)
	suite.Len(suite.mailRequests, 1)

	_, ok := suite.lastSentAt(coursePhaseID, assessmentType.Self)
	suite.False(ok)
}

func (suite *EvaluationReminderSendTestSuite) TestReadFailureAbortsBeforeMailing() {
	_, err := suite.service.conn.Exec(suite.ctx, "ALTER TABLE evaluation_reminder RENAME TO evaluation_reminder_unavailable")
	suite.Require().NoError(err)
	suite.T().Cleanup(func() {
		_, err := suite.service.conn.Exec(suite.ctx, "ALTER TABLE evaluation_reminder_unavailable RENAME TO evaluation_reminder")
		suite.Require().NoError(err)
	})

	_, err = suite.service.SendEvaluationReminderManualTrigger(suite.ctx, "Bearer token", uuid.New(), assessmentType.Self)
	suite.Require().Error(err)
	suite.Empty(suite.mailRequests)
}

func TestEvaluationReminderSendTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluationReminderSendTestSuite))
}

func newReminderSendTestService(recipients reminderRecipientsResolver) *CoursePhaseConfigService {
	service := NewCoursePhaseConfigService(db.Queries{}, nil, nil)
	service.getEvaluationReminderRecipients = recipients
	return service
}

func TestSendEvaluationReminderManualTriggerDeadlineNotPassed(t *testing.T) {
	deadline := time.Date(2026, time.January, 20, 15, 0, 0, 0, time.UTC)
	service := newReminderSendTestService(func(
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
	})

	_, err := service.SendEvaluationReminderManualTrigger(context.Background(), "Bearer token", uuid.New(), assessmentType.Self)
	require.ErrorIs(t, err, ErrReminderDeadlineNotPassed)
}

func TestSendEvaluationReminderManualTriggerTemplateIncomplete(t *testing.T) {
	oldGetCoreCoursePhaseFn := getCoreCoursePhaseFn
	t.Cleanup(func() {
		getCoreCoursePhaseFn = oldGetCoreCoursePhaseFn
	})

	service := newReminderSendTestService(func(
		ctx context.Context,
		authHeader string,
		coursePhaseID uuid.UUID,
		evaluationType assessmentType.AssessmentType,
	) (coursePhaseConfigDTO.EvaluationReminderRecipients, error) {
		return coursePhaseConfigDTO.EvaluationReminderRecipients{
			EvaluationEnabled: true,
			DeadlinePassed:    true,
		}, nil
	})

	getCoreCoursePhaseFn = func(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (coreCoursePhaseResponse, error) {
		return coreCoursePhaseResponse{
			ID:             coursePhaseID,
			Name:           "Assessment Phase",
			RestrictedData: map[string]any{},
		}, nil
	}

	_, err := service.SendEvaluationReminderManualTrigger(context.Background(), "Bearer token", uuid.New(), assessmentType.Self)
	require.ErrorIs(t, err, ErrReminderTemplateIncomplete)
}
