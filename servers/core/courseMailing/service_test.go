package courseMailing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/courseMailing/courseMailingDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testCourseID   = "11111111-1111-1111-1111-111111111111"
	testPhaseID    = "33333333-3333-3333-3333-333333333333"
	testEmptyPhase = "cccccccc-cccc-cccc-cccc-cccccccccccc"
)

type capturedMail struct {
	settings  mailingDTO.CourseMailingSettings
	recipient string
	subject   string
	body      string
}

type CourseMailingServiceTestSuite struct {
	suite.Suite
	ctx        context.Context
	cleanup    func()
	service    *CourseMailingService
	courseID   uuid.UUID
	phaseID    uuid.UUID
	emptyPhase uuid.UUID
	actor      courseMailingDTO.Actor

	oldSendMailFn   func(mailingDTO.CourseMailingSettings, string, string, string) error
	oldRunSendAsync func(func())
	oldNowFn        func() time.Time
	captured        []capturedMail
}

func (suite *CourseMailingServiceTestSuite) SetupSuite() {
	if testing.Short() {
		suite.T().Skip("skipping db-backed course mailing tests in short mode")
	}
	defer func() {
		if r := recover(); r != nil {
			suite.T().Skipf("skipping db-backed course mailing tests: %v", r)
		}
	}()

	suite.ctx = context.Background()
	suite.courseID = uuid.MustParse(testCourseID)
	suite.phaseID = uuid.MustParse(testPhaseID)
	suite.emptyPhase = uuid.MustParse(testEmptyPhase)
	suite.actor = courseMailingDTO.Actor{ID: "user-1", Email: "lecturer@example.com", Name: "Lena Lecturer"}

	testDB, cleanup, err := testutils.SetupTestDB(
		suite.ctx,
		"../database_dumps/course_mail_campaign_test.sql",
		func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) },
	)
	if err != nil {
		suite.T().Skipf("skipping db-backed course mailing tests: %v", err)
	}
	suite.cleanup = cleanup

	suite.service = &CourseMailingService{
		queries:   *testDB.Queries,
		conn:      testDB.Conn,
		clientURL: "https://prompt.example.com",
	}
	CourseMailingServiceSingleton = suite.service

	suite.oldSendMailFn = sendMailFn
	suite.oldRunSendAsync = runSendAsync
	suite.oldNowFn = nowFn
}

func (suite *CourseMailingServiceTestSuite) TearDownSuite() {
	sendMailFn = suite.oldSendMailFn
	runSendAsync = suite.oldRunSendAsync
	nowFn = suite.oldNowFn
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CourseMailingServiceTestSuite) SetupTest() {
	// Run sends synchronously and capture mails by default.
	runSendAsync = func(fn func()) { fn() }
	suite.captured = nil
	sendMailFn = func(settings mailingDTO.CourseMailingSettings, recipient, subject, body string) error {
		suite.captured = append(suite.captured, capturedMail{settings: settings, recipient: recipient, subject: subject, body: body})
		return nil
	}
	nowFn = func() time.Time { return time.Date(2026, time.July, 28, 12, 0, 0, 0, time.UTC) }
}

func (suite *CourseMailingServiceTestSuite) newCampaign(name, subject, body string, phase *uuid.UUID, statuses []string) db.MailCampaign {
	campaign, err := suite.service.CreateCampaign(suite.ctx, suite.courseID, suite.actor, courseMailingDTO.MailCampaignRequest{
		Name:                name,
		Subject:             subject,
		Body:                body,
		TargetCoursePhaseID: phase,
		TargetPassStatuses:  statuses,
	})
	suite.Require().NoError(err)
	return campaign
}

func (suite *CourseMailingServiceTestSuite) statusOf(campaignID uuid.UUID) db.MailCampaignStatus {
	base, err := suite.service.getBase(suite.ctx, suite.courseID, campaignID)
	suite.Require().NoError(err)
	return base.Status
}

func (suite *CourseMailingServiceTestSuite) TestCRUDLifecycle() {
	created := suite.newCampaign("My Campaign", "Hi {{firstName}}", "Hello", &suite.phaseID, []string{"passed"})
	suite.Equal(db.MailCampaignStatusDraft, created.Status)
	suite.Equal("user-1", created.CreatedByID)

	detail, err := suite.service.GetCampaignDetail(suite.ctx, suite.courseID, created.ID)
	suite.Require().NoError(err)
	suite.Equal("My Campaign", detail.Name)
	suite.Len(detail.Recipients, 0)

	updated, err := suite.service.UpdateCampaign(suite.ctx, suite.courseID, created.ID, courseMailingDTO.Actor{ID: "user-2", Email: "e@x.de", Name: "Ed Editor"}, courseMailingDTO.MailCampaignRequest{
		Name:                "Renamed",
		Subject:             "Hi",
		Body:                "Body",
		TargetCoursePhaseID: &suite.phaseID,
		TargetPassStatuses:  []string{"failed"},
	})
	suite.Require().NoError(err)
	suite.Equal("Renamed", updated.Name)
	suite.Equal("user-2", updated.UpdatedByID)

	list, err := suite.service.ListCampaigns(suite.ctx, suite.courseID)
	suite.Require().NoError(err)
	suite.GreaterOrEqual(len(list), 1)

	suite.Require().NoError(suite.service.DeleteCampaign(suite.ctx, suite.courseID, created.ID))
	_, err = suite.service.GetCampaignDetail(suite.ctx, suite.courseID, created.ID)
	suite.ErrorIs(err, ErrNotFound)
}

func (suite *CourseMailingServiceTestSuite) TestCopyCreatesDraft() {
	original := suite.newCampaign("Original", "Sub", "Body", &suite.phaseID, []string{"failed"})
	copied, err := suite.service.CopyCampaign(suite.ctx, suite.courseID, original.ID, suite.actor)
	suite.Require().NoError(err)

	suite.NotEqual(original.ID, copied.ID)
	suite.Equal(db.MailCampaignStatusDraft, copied.Status)
	suite.True(strings.HasSuffix(copied.Name, "(Copy)"))
	suite.False(copied.SentAt.Valid)
	suite.Equal(original.Subject, copied.Subject)
}

func (suite *CourseMailingServiceTestSuite) TestCreateRejectsForeignPhase() {
	foreignPhase := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	_, err := suite.service.CreateCampaign(suite.ctx, suite.courseID, suite.actor, courseMailingDTO.MailCampaignRequest{
		Name:                "Cross-course",
		TargetCoursePhaseID: &foreignPhase,
		TargetPassStatuses:  []string{"passed"},
	})
	suite.ErrorIs(err, ErrValidation)
}

func (suite *CourseMailingServiceTestSuite) TestUpdateRejectsForeignPhase() {
	campaign := suite.newCampaign("Campaign", "Sub", "Body", &suite.phaseID, []string{"passed"})
	foreignPhase := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	_, err := suite.service.UpdateCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor, courseMailingDTO.MailCampaignRequest{
		Name:                "Campaign",
		TargetCoursePhaseID: &foreignPhase,
		TargetPassStatuses:  []string{"passed"},
	})
	suite.ErrorIs(err, ErrValidation)
}

func (suite *CourseMailingServiceTestSuite) TestPreviewRecipientsByStatus() {
	cases := []struct {
		statuses []string
		expected int
	}{
		{[]string{"passed"}, 2},
		{[]string{"failed"}, 1},
		{[]string{"not_assessed"}, 1},
		{[]string{"all"}, 4},
		{[]string{"passed", "failed"}, 3},
	}
	for _, tc := range cases {
		campaign := suite.newCampaign("Preview", "Sub", "Body", &suite.phaseID, tc.statuses)
		preview, err := suite.service.PreviewRecipients(suite.ctx, suite.courseID, campaign.ID)
		suite.Require().NoError(err)
		suite.Equalf(tc.expected, preview.Count, "statuses %v", tc.statuses)
	}
}

func (suite *CourseMailingServiceTestSuite) TestSendPartialFailureMissingEmail() {
	// passed -> Alice (email) + Dan (no email). Dan must fail with a missing-email error.
	campaign := suite.newCampaign("Send", "Hi {{firstName}}", "Hello {{firstName}} <{{email}}>", &suite.phaseID, []string{"passed"})

	count, err := suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)
	suite.Equal(2, count)

	suite.Equal(db.MailCampaignStatusPartiallyFailed, suite.statusOf(campaign.ID))

	detail, err := suite.service.GetCampaignDetail(suite.ctx, suite.courseID, campaign.ID)
	suite.Require().NoError(err)
	suite.Equal(int64(1), detail.SentCount)
	suite.Equal(int64(1), detail.FailedCount)

	// Alice's rendered mail: placeholders resolved, none left over.
	suite.Require().Len(suite.captured, 1)
	suite.Contains(suite.captured[0].subject, "Alice")
	suite.NotContains(suite.captured[0].body, "{{")

	var danFailed bool
	for _, r := range detail.Recipients {
		if r.FirstName == "Dan" {
			suite.Equal(string(db.MailRecipientStatusFailed), r.Status)
			suite.Contains(r.ErrorMessage, "missing email")
			danFailed = true
		}
	}
	suite.True(danFailed)
}

func (suite *CourseMailingServiceTestSuite) TestSendAllSuccessSetsSentMeta() {
	campaign := suite.newCampaign("Send", "Hi {{firstName}}", "Body", &suite.phaseID, []string{"failed"})
	count, err := suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)
	suite.Equal(1, count)

	base, err := suite.service.getBase(suite.ctx, suite.courseID, campaign.ID)
	suite.Require().NoError(err)
	suite.Equal(db.MailCampaignStatusSent, base.Status)
	suite.True(base.SentAt.Valid)
	suite.Equal("user-1", base.SentByID.String)
	suite.Equal("lecturer@example.com", base.SentByEmail.String)
}

func (suite *CourseMailingServiceTestSuite) TestSendSMTPErrorMarksFailed() {
	sendMailFn = func(mailingDTO.CourseMailingSettings, string, string, string) error {
		return errors.New("smtp down")
	}
	campaign := suite.newCampaign("Send", "Hi", "Body", &suite.phaseID, []string{"failed"})
	_, err := suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)
	suite.Equal(db.MailCampaignStatusFailed, suite.statusOf(campaign.ID))
}

func (suite *CourseMailingServiceTestSuite) TestSendZeroRecipients() {
	campaign := suite.newCampaign("Send", "Hi", "Body", &suite.emptyPhase, []string{"passed"})
	_, err := suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.ErrorIs(err, ErrNoRecipients)
	// Status untouched (no state change before the zero-recipient guard).
	suite.Equal(db.MailCampaignStatusDraft, suite.statusOf(campaign.ID))
}

func (suite *CourseMailingServiceTestSuite) TestConcurrentSendConflict() {
	campaign := suite.newCampaign("Send", "Hi", "Body", &suite.phaseID, []string{"failed"})
	// Simulate an in-flight send by flipping the status to sending.
	_, err := suite.service.queries.TrySetMailCampaignSending(suite.ctx, db.TrySetMailCampaignSendingParams{ID: campaign.ID, CourseID: suite.courseID})
	suite.Require().NoError(err)

	_, err = suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.ErrorIs(err, ErrSendInProgress)
}

func (suite *CourseMailingServiceTestSuite) TestSendRejectsAlreadySentCampaign() {
	campaign := suite.newCampaign("Send", "Hi", "Body", &suite.phaseID, []string{"failed"})
	_, err := suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)
	suite.Equal(db.MailCampaignStatusSent, suite.statusOf(campaign.ID))

	sentCount := len(suite.captured)
	_, err = suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.ErrorIs(err, ErrValidation)
	suite.Len(suite.captured, sentCount, "no additional mail must be sent for an already-sent campaign")
}

func (suite *CourseMailingServiceTestSuite) TestResendFailedRecovers() {
	// First send fails for everyone.
	sendMailFn = func(mailingDTO.CourseMailingSettings, string, string, string) error {
		return errors.New("smtp down")
	}
	campaign := suite.newCampaign("Send", "Hi", "Body", &suite.phaseID, []string{"failed"})
	_, err := suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)
	suite.Equal(db.MailCampaignStatusFailed, suite.statusOf(campaign.ID))

	// Resend succeeds.
	sendMailFn = func(mailingDTO.CourseMailingSettings, string, string, string) error { return nil }
	count, err := suite.service.ResendFailed(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)
	suite.Equal(1, count)
	suite.Equal(db.MailCampaignStatusSent, suite.statusOf(campaign.ID))
}

func (suite *CourseMailingServiceTestSuite) TestResendFailedSkipsRecipientsNoLongerMatchingStatus() {
	sendMailFn = func(mailingDTO.CourseMailingSettings, string, string, string) error {
		return errors.New("smtp down")
	}
	campaign := suite.newCampaign("Send", "Hi", "Body", &suite.phaseID, []string{"failed"})
	_, err := suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)
	suite.Equal(db.MailCampaignStatusFailed, suite.statusOf(campaign.ID))

	// Bob (the only "failed" participant) is reassessed as "passed" before the resend.
	// Other tests share this fixture DB, so restore his status once this test is done.
	bobParticipation := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	_, err = suite.service.queries.UpdateCoursePhasePassStatus(suite.ctx, db.UpdateCoursePhasePassStatusParams{
		PassStatus:            db.PassStatusPassed,
		CourseParticipationID: []uuid.UUID{bobParticipation},
		CoursePhaseID:         suite.phaseID,
	})
	suite.Require().NoError(err)
	defer func() {
		_, _ = suite.service.queries.UpdateCoursePhasePassStatus(suite.ctx, db.UpdateCoursePhasePassStatusParams{
			PassStatus:            db.PassStatusFailed,
			CourseParticipationID: []uuid.UUID{bobParticipation},
			CoursePhaseID:         suite.phaseID,
		})
	}()

	sendMailFn = func(mailingDTO.CourseMailingSettings, string, string, string) error { return nil }
	_, err = suite.service.ResendFailed(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.ErrorIs(err, ErrValidation)
	suite.Empty(suite.captured, "no mail must be sent to a recipient outside the campaign's target statuses")
}

func (suite *CourseMailingServiceTestSuite) TestResendNoFailedRecipients() {
	campaign := suite.newCampaign("Send", "Hi", "Body", &suite.phaseID, []string{"failed"})
	_, err := suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)

	_, err = suite.service.ResendFailed(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.ErrorIs(err, ErrValidation)
}

func (suite *CourseMailingServiceTestSuite) TestCCBCCSentOnce() {
	campaign, err := suite.service.CreateCampaign(suite.ctx, suite.courseID, suite.actor, courseMailingDTO.MailCampaignRequest{
		Name:                "CC Campaign",
		Subject:             "Hi",
		Body:                "Body",
		TargetCoursePhaseID: &suite.phaseID,
		TargetPassStatuses:  []string{"failed"}, // one student (Bob)
		CcOverride:          []courseMailingDTO.MailItem{{Name: "Archive", Email: "archive@example.com"}},
		BccOverride:         []courseMailingDTO.MailItem{{Name: "Audit", Email: "audit@example.com"}},
	})
	suite.Require().NoError(err)

	_, err = suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)

	// One student mail + exactly one archive copy.
	suite.Require().Len(suite.captured, 2)

	var studentCalls, archiveCalls int
	for _, c := range suite.captured {
		if c.recipient == "bob@example.com" {
			studentCalls++
			suite.Empty(c.settings.CC, "student mail must not carry CC")
			suite.Empty(c.settings.BCC, "student mail must not carry BCC")
		} else {
			archiveCalls++
			suite.Equal("replyto@example.com", c.recipient)
			suite.Contains(addressEmails(c.settings.CC), "archive@example.com")
			suite.Contains(addressEmails(c.settings.BCC), "audit@example.com")
		}
	}
	suite.Equal(1, studentCalls)
	suite.Equal(1, archiveCalls)
}

func (suite *CourseMailingServiceTestSuite) TestArchiveCopyLeavesStudentPlaceholdersUntouched() {
	campaign, err := suite.service.CreateCampaign(suite.ctx, suite.courseID, suite.actor, courseMailingDTO.MailCampaignRequest{
		Name:                "CC Campaign",
		Subject:             "Hi",
		Body:                "Dear {{firstName}}, degree {{studyDegree}}",
		TargetCoursePhaseID: &suite.phaseID,
		TargetPassStatuses:  []string{"failed"}, // one student (Bob)
		CcOverride:          []courseMailingDTO.MailItem{{Name: "Archive", Email: "archive@example.com"}},
	})
	suite.Require().NoError(err)

	_, err = suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)

	var archiveBody, studentBody string
	for _, c := range suite.captured {
		if c.recipient == "bob@example.com" {
			studentBody = c.body
		} else {
			archiveBody = c.body
		}
	}
	suite.Contains(studentBody, "Dear Bob, degree Master")
	suite.Contains(archiveBody, "{{firstName}}", "the archive copy has no single recipient, so student placeholders must stay unresolved")
	suite.Contains(archiveBody, "{{studyDegree}}")
	suite.NotContains(archiveBody, "Unknown")
}

func (suite *CourseMailingServiceTestSuite) TestCreateRejectsInvalidOverrideEmail() {
	_, err := suite.service.CreateCampaign(suite.ctx, suite.courseID, suite.actor, courseMailingDTO.MailCampaignRequest{
		Name:                "Bad CC",
		TargetCoursePhaseID: &suite.phaseID,
		TargetPassStatuses:  []string{"passed"},
		CcOverride:          []courseMailingDTO.MailItem{{Name: "Broken", Email: "not-an-email"}},
	})
	suite.ErrorIs(err, ErrValidation)
}

func (suite *CourseMailingServiceTestSuite) TestCreateRejectsEmptyNameAtServiceBoundary() {
	// The service validates independently of the HTTP router.
	_, err := suite.service.CreateCampaign(suite.ctx, suite.courseID, suite.actor, courseMailingDTO.MailCampaignRequest{
		Name:                "   ",
		TargetCoursePhaseID: &suite.phaseID,
		TargetPassStatuses:  []string{"passed"},
	})
	suite.ErrorIs(err, ErrValidation)
}

func (suite *CourseMailingServiceTestSuite) TestSendRejectsSubjectWithLineBreak() {
	campaign := suite.newCampaign("Injection", "Hi\r\nBcc: attacker@example.com", "Body", &suite.phaseID, []string{"failed"})
	_, err := suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.ErrorIs(err, ErrValidation)
	suite.Empty(suite.captured, "no mail must be sent for an invalid subject")
}

func (suite *CourseMailingServiceTestSuite) TestArchiveCopyRendersCoursePhaseLink() {
	campaign, err := suite.service.CreateCampaign(suite.ctx, suite.courseID, suite.actor, courseMailingDTO.MailCampaignRequest{
		Name:                "Link Campaign",
		Subject:             "Hi",
		Body:                "Open {{coursePhaseLink}}",
		TargetCoursePhaseID: &suite.phaseID,
		TargetPassStatuses:  []string{"failed"}, // one student (Bob)
		CcOverride:          []courseMailingDTO.MailItem{{Name: "Archive", Email: "archive@example.com"}},
	})
	suite.Require().NoError(err)

	_, err = suite.service.SendCampaign(suite.ctx, suite.courseID, campaign.ID, suite.actor)
	suite.Require().NoError(err)

	var archiveBody string
	for _, c := range suite.captured {
		if c.recipient == "replyto@example.com" {
			archiveBody = c.body
		}
	}
	suite.Contains(archiveBody, "https://prompt.example.com/management/course/")
	suite.NotContains(archiveBody, "{{coursePhaseLink}}")
}

func addressEmails(addresses []mail.Address) []string {
	out := make([]string, 0, len(addresses))
	for _, a := range addresses {
		out = append(out, a.Address)
	}
	return out
}

func TestComputeCampaignStatus(t *testing.T) {
	tests := []struct {
		name       string
		recipients []db.MailCampaignRecipient
		want       db.MailCampaignStatus
	}{
		{"no recipients", nil, db.MailCampaignStatusFailed},
		{"all pending", []db.MailCampaignRecipient{
			{Status: db.MailRecipientStatusPending},
		}, db.MailCampaignStatusFailed},
		{"all failed", []db.MailCampaignRecipient{
			{Status: db.MailRecipientStatusFailed},
		}, db.MailCampaignStatusFailed},
		{"all sent", []db.MailCampaignRecipient{
			{Status: db.MailRecipientStatusSent},
		}, db.MailCampaignStatusSent},
		{"mixed sent and failed", []db.MailCampaignRecipient{
			{Status: db.MailRecipientStatusSent},
			{Status: db.MailRecipientStatusFailed},
		}, db.MailCampaignStatusPartiallyFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, computeCampaignStatus(tt.recipients))
		})
	}
}

// --- HTTP role-gating tests -------------------------------------------------

func enforceRoles(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = true
	}
	return func(c *gin.Context) {
		rolesVal, _ := c.Get("userRoles")
		roles, _ := rolesVal.(map[string]bool)
		for r := range roles {
			if allowedSet[r] {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

func buildEngine(roles []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	setupCourseMailingRouter(api, func() gin.HandlerFunc {
		return testutils.MockAuthMiddleware(roles)
	}, enforceRoles)
	return router
}

func (suite *CourseMailingServiceTestSuite) TestEditorCannotSend() {
	campaign := suite.newCampaign("Send", "Hi", "Body", &suite.phaseID, []string{"failed"})
	router := buildEngine([]string{permissionValidation.CourseEditor})

	req := httptest.NewRequest(http.MethodPost, "/api/courses/"+testCourseID+"/mail-campaigns/"+campaign.ID.String()+"/send", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	suite.Equal(http.StatusForbidden, resp.Code)
}

func (suite *CourseMailingServiceTestSuite) TestLecturerCanSend() {
	campaign := suite.newCampaign("Send", "Hi", "Body", &suite.phaseID, []string{"failed"})
	router := buildEngine([]string{permissionValidation.CourseLecturer})

	req := httptest.NewRequest(http.MethodPost, "/api/courses/"+testCourseID+"/mail-campaigns/"+campaign.ID.String()+"/send", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	suite.Equal(http.StatusAccepted, resp.Code)
	suite.Equal(db.MailCampaignStatusSent, suite.statusOf(campaign.ID))
}

func TestCourseMailingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CourseMailingServiceTestSuite))
}

func TestExpandStatuses(t *testing.T) {
	got, err := expandStatuses([]string{"all"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 statuses from 'all', got %d", len(got))
	}

	if _, err := expandStatuses([]string{"bogus"}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error for bogus status, got %v", err)
	}
	if _, err := expandStatuses(nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error for empty statuses, got %v", err)
	}
}
