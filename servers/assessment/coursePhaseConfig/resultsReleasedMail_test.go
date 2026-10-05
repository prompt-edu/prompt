package coursePhaseConfig

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/suite"
)

type ResultsReleasedMailTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	service *CoursePhaseConfigService

	restoreStubs   func()
	mailingEnabled bool
	participants   []coursePhaseConfigDTO.AssessmentParticipationWithStudent
	mailRequests   []coreManualMailRequest
	failEmails     []string
	skipEmails     []string
	coreMailErr    error
}

func (suite *ResultsReleasedMailTestSuite) SetupSuite() {
	if testing.Short() {
		suite.T().Skip("skipping db-backed results released mail tests in short mode")
	}
	defer func() {
		if r := recover(); r != nil {
			suite.T().Skipf("skipping db-backed results released mail tests: %v", r)
		}
	}()

	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/coursePhaseConfig.sql")
	if err != nil {
		suite.T().Skipf("skipping db-backed results released mail tests: %v", err)
	}
	suite.cleanup = cleanup
	suite.service = NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn, nil)
}

func (suite *ResultsReleasedMailTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ResultsReleasedMailTestSuite) SetupTest() {
	suite.mailingEnabled = true
	suite.participants = nil
	suite.mailRequests = nil
	suite.failEmails = nil
	suite.skipEmails = nil
	suite.coreMailErr = nil

	oldGetCoreCoursePhaseFn := getCoreCoursePhaseFn
	oldGetParticipationsForCoursePhaseFn := getParticipationsForCoursePhaseFn
	oldSendManualMailFn := sendManualMailFn
	suite.restoreStubs = func() {
		getCoreCoursePhaseFn = oldGetCoreCoursePhaseFn
		getParticipationsForCoursePhaseFn = oldGetParticipationsForCoursePhaseFn
		sendManualMailFn = oldSendManualMailFn
	}

	getCoreCoursePhaseFn = func(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (coreCoursePhaseResponse, error) {
		return coreCoursePhaseResponse{
			ID:   coursePhaseID,
			Name: "Assessment Phase",
			RestrictedData: map[string]any{
				"mailingSettings": map[string]any{
					"resultsReleasedMail": map[string]any{
						"subject":       "Results for {{coursePhaseName}}",
						"content":       "Hi {{firstName}}, your results are available.",
						"sendOnRelease": suite.mailingEnabled,
					},
				},
			},
		}, nil
	}
	getParticipationsForCoursePhaseFn = func(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) ([]coursePhaseConfigDTO.AssessmentParticipationWithStudent, error) {
		return append([]coursePhaseConfigDTO.AssessmentParticipationWithStudent(nil), suite.participants...), nil
	}
	sendManualMailFn = func(ctx context.Context, authHeader string, coursePhaseID uuid.UUID, request coreManualMailRequest) (coreManualMailReport, error) {
		suite.mailRequests = append(suite.mailRequests, request)
		if suite.coreMailErr != nil {
			return coreManualMailReport{}, suite.coreMailErr
		}
		report := coreManualMailReport{RequestedRecipients: len(request.RecipientCourseParticipationIDs)}
		for _, participant := range suite.participants {
			if !slices.Contains(request.RecipientCourseParticipationIDs, participant.CourseParticipationID) {
				continue
			}
			if slices.Contains(suite.skipEmails, participant.Student.Email) {
				continue
			}
			if slices.Contains(suite.failEmails, participant.Student.Email) {
				report.FailedEmails = append(report.FailedEmails, participant.Student.Email)
			} else {
				report.SuccessfulEmails = append(report.SuccessfulEmails, participant.Student.Email)
			}
		}
		return report, nil
	}
}

func (suite *ResultsReleasedMailTestSuite) TearDownTest() {
	suite.restoreStubs()
}

func (suite *ResultsReleasedMailTestSuite) createPhase(assessmentEnabled bool) uuid.UUID {
	coursePhaseID := uuid.New()
	_, err := suite.service.conn.Exec(suite.ctx,
		"INSERT INTO course_phase_config (assessment_schema_id, course_phase_id, assessment_enabled, results_released) VALUES ($1, $2, $3, true)",
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), coursePhaseID, assessmentEnabled)
	suite.Require().NoError(err)
	return coursePhaseID
}

func (suite *ResultsReleasedMailTestSuite) addParticipant(coursePhaseID uuid.UUID, email string, assessmentFinal bool) uuid.UUID {
	courseParticipationID := uuid.New()
	suite.participants = append(suite.participants, coursePhaseConfigDTO.AssessmentParticipationWithStudent{
		CoursePhaseParticipationWithStudent: promptTypes.CoursePhaseParticipationWithStudent{
			CoursePhaseID:         coursePhaseID,
			CourseParticipationID: courseParticipationID,
			Student:               promptTypes.Student{Email: email},
		},
	})
	_, err := suite.service.conn.Exec(suite.ctx,
		"INSERT INTO assessment_completion (course_participation_id, course_phase_id, completed_at, author, completed) VALUES ($1, $2, $3, 'tutor', $4)",
		courseParticipationID, coursePhaseID, time.Now(), assessmentFinal)
	suite.Require().NoError(err)
	return courseParticipationID
}

func (suite *ResultsReleasedMailTestSuite) send(coursePhaseID uuid.UUID) *coursePhaseConfigDTO.ResultsReleasedMailReport {
	report, err := suite.service.SendResultsReleasedMail(suite.ctx, "Bearer token", coursePhaseID)
	suite.Require().NoError(err)
	return report
}

func (suite *ResultsReleasedMailTestSuite) TestSendsNothingWhenDisabled() {
	coursePhaseID := suite.createPhase(true)
	suite.addParticipant(coursePhaseID, "alice@example.com", true)
	suite.mailingEnabled = false

	suite.Nil(suite.send(coursePhaseID))
	suite.Empty(suite.mailRequests)
}

func (suite *ResultsReleasedMailTestSuite) TestMailsFinalAssessmentsOnlyOnce() {
	coursePhaseID := suite.createPhase(true)
	alice := suite.addParticipant(coursePhaseID, "alice@example.com", true)
	suite.addParticipant(coursePhaseID, "bob@example.com", false)

	report := suite.send(coursePhaseID)
	suite.Equal([]string{"alice@example.com"}, report.SuccessfulEmails)
	suite.Require().Len(suite.mailRequests, 1)
	suite.Equal([]uuid.UUID{alice}, suite.mailRequests[0].RecipientCourseParticipationIDs)
	suite.Equal("Assessment Phase", suite.mailRequests[0].AdditionalPlaceholders["coursePhaseName"])

	report = suite.send(coursePhaseID)
	suite.Empty(report.SuccessfulEmails)
	suite.Len(suite.mailRequests, 1)
}

func (suite *ResultsReleasedMailTestSuite) TestMailsEveryParticipantOfEvaluationOnlyPhase() {
	coursePhaseID := suite.createPhase(false)
	suite.addParticipant(coursePhaseID, "alice@example.com", false)
	suite.addParticipant(coursePhaseID, "bob@example.com", false)

	report := suite.send(coursePhaseID)
	suite.ElementsMatch([]string{"alice@example.com", "bob@example.com"}, report.SuccessfulEmails)
}

func (suite *ResultsReleasedMailTestSuite) TestRetriesFailedRecipientOnNextRelease() {
	coursePhaseID := suite.createPhase(true)
	suite.addParticipant(coursePhaseID, "alice@example.com", true)
	bob := suite.addParticipant(coursePhaseID, "bob@example.com", true)
	suite.failEmails = []string{"bob@example.com"}

	report := suite.send(coursePhaseID)
	suite.Equal([]string{"alice@example.com"}, report.SuccessfulEmails)
	suite.Equal([]string{"bob@example.com"}, report.FailedEmails)

	suite.failEmails = nil
	report = suite.send(coursePhaseID)
	suite.Equal([]string{"bob@example.com"}, report.SuccessfulEmails)
	suite.Equal([]uuid.UUID{bob}, suite.mailRequests[1].RecipientCourseParticipationIDs)
}

func (suite *ResultsReleasedMailTestSuite) TestRetriesRecipientCoreSkipped() {
	coursePhaseID := suite.createPhase(true)
	suite.addParticipant(coursePhaseID, "alice@example.com", true)
	suite.addParticipant(coursePhaseID, "bob@example.com", true)
	suite.skipEmails = []string{"bob@example.com"}

	report := suite.send(coursePhaseID)
	suite.Equal([]string{"alice@example.com"}, report.SuccessfulEmails)

	suite.skipEmails = nil
	report = suite.send(coursePhaseID)
	suite.Equal([]string{"bob@example.com"}, report.SuccessfulEmails)
}

func (suite *ResultsReleasedMailTestSuite) TestReleasesAllClaimsWhenCoreRejects() {
	coursePhaseID := suite.createPhase(true)
	suite.addParticipant(coursePhaseID, "alice@example.com", true)
	suite.coreMailErr = fmt.Errorf("%w with status 500: boom", errCoreRejectedMail)

	_, err := suite.service.SendResultsReleasedMail(suite.ctx, "Bearer token", coursePhaseID)
	suite.Error(err)

	suite.coreMailErr = nil
	report := suite.send(coursePhaseID)
	suite.Equal([]string{"alice@example.com"}, report.SuccessfulEmails)
}

func (suite *ResultsReleasedMailTestSuite) TestKeepsClaimsWhenCoreOutcomeIsUnknown() {
	coursePhaseID := suite.createPhase(true)
	suite.addParticipant(coursePhaseID, "alice@example.com", true)
	suite.coreMailErr = fmt.Errorf("failed to send manual mails via core: %w", context.DeadlineExceeded)

	_, err := suite.service.SendResultsReleasedMail(suite.ctx, "Bearer token", coursePhaseID)
	suite.ErrorIs(err, errResultsMailOutcomeUnknown)

	suite.coreMailErr = nil
	report := suite.send(coursePhaseID)
	suite.Empty(report.SuccessfulEmails)
	suite.Len(suite.mailRequests, 1)
}

func TestSendManualMailMarksOnlyCoreRejections(t *testing.T) {
	status, body := http.StatusInternalServerError, "boom"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()
	t.Setenv("SERVER_CORE_HOST", server.URL)

	_, err := sendManualMail(context.Background(), "", uuid.New(), coreManualMailRequest{})
	if !errors.Is(err, errCoreRejectedMail) {
		t.Fatalf("expected a core rejection for status 500, got %v", err)
	}

	status, body = http.StatusOK, "not json"
	_, err = sendManualMail(context.Background(), "", uuid.New(), coreManualMailRequest{})
	if err == nil || errors.Is(err, errCoreRejectedMail) {
		t.Fatalf("expected an unparseable 200 response not to count as a rejection, got %v", err)
	}
}

func TestResultsReleasedMailTestSuite(t *testing.T) {
	suite.Run(t, new(ResultsReleasedMailTestSuite))
}
