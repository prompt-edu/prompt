package privacy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/ai/calls"
	"github.com/prompt-edu/prompt/servers/ai/calls/callDTO"
	"github.com/prompt-edu/prompt/servers/ai/feature"
	"github.com/prompt-edu/prompt/servers/ai/privacy/privacyDTO"
	"github.com/prompt-edu/prompt/servers/ai/testutils"
	"github.com/stretchr/testify/suite"
)

const lowRiskFeature = "example.course_editing"

type PrivacySuite struct {
	suite.Suite
	ctx     context.Context
	db      *testutils.TestDB
	cleanup func()
	calls   *calls.Service
	service *Service

	lecturer uuid.UUID
	student  uuid.UUID
	other    uuid.UUID
}

func (s *PrivacySuite) SetupTest() {
	s.ctx = context.Background()
	s.T().Setenv("AI_ENCRYPTION_KEY", "ZTJlLWFpLXRlc3Qta2V5LW5vdC1hLXJlYWwtc2VjcmU=")
	testDB, cleanup, err := testutils.SetupTestDB(s.ctx)
	s.Require().NoError(err)
	s.db, s.cleanup = testDB, cleanup
	s.calls = calls.NewService(testDB.Queries, testDB.Conn, "logos.test", "test")
	s.service = NewService(testDB.Queries, testDB.Conn)
	s.service.policyOf = func(name string) feature.Policy {
		if name == lowRiskFeature {
			return feature.Policy{ContentRetentionDays: 30}
		}
		return feature.PolicyOf(name)
	}
	s.lecturer, s.student, s.other = uuid.New(), uuid.New(), uuid.New()
}

func (s *PrivacySuite) TearDownTest() { s.cleanup() }

func TestPrivacySuite(t *testing.T) { suite.Run(t, new(PrivacySuite)) }

func (s *PrivacySuite) completedCall(featureName string, subjects ...uuid.UUID) uuid.UUID {
	callID, err := s.calls.Begin(s.ctx, calls.Request{
		CoursePhaseID: uuid.New(),
		ActorID:       s.lecturer.String(),
		ActorRole:     "Lecturer",
		Feature:       featureName,
		Model:         "m1",
		Subjects:      subjects,
		Body:          []byte(`{"messages":[{"role":"user","content":"About the student"}]}`),
	})
	s.Require().NoError(err)
	s.Require().NoError(s.calls.Finish(s.ctx, callID, calls.Completion{
		Outcome:  calls.OutcomeSuccess,
		Response: []byte(`{"model":"m1","choices":[{"message":{"content":"Suggestion"},"finish_reason":"stop"}]}`),
	}))
	return callID
}

func (s *PrivacySuite) ginContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil).WithContext(s.ctx)
	return c
}

func (s *PrivacySuite) count(query string, args ...any) int {
	var count int
	s.Require().NoError(s.db.Conn.QueryRow(s.ctx, query, args...).Scan(&count))
	return count
}

func (s *PrivacySuite) TestExportCoversSubjectAndActor() {
	aboutStudent := s.completedCall("assessment.action_item_suggestions", s.student)
	aboutBoth := s.completedCall("assessment.action_item_suggestions", s.student, s.other)

	asStudent, err := s.service.exportCalls(s.ctx, uuid.NewString(), []uuid.UUID{s.student})
	s.Require().NoError(err)
	s.Require().Len(asStudent, 2)
	s.Equal(aboutStudent, asStudent[0].ID)
	s.False(asStudent[0].MadeBySubject)
	s.Equal(callDTO.ContentAvailable, asStudent[0].ContentState)
	s.Contains(string(asStudent[0].Request), "About the student")
	s.Equal("Suggestion", asStudent[0].Response)
	s.Equal(aboutBoth, asStudent[1].ID)
	s.Equal(privacyDTO.ContentWithheld, asStudent[1].ContentState, "content about other people stays out of the export")
	s.Nil(asStudent[1].Request)

	asLecturer, err := s.service.exportCalls(s.ctx, s.lecturer.String(), nil)
	s.Require().NoError(err)
	s.Len(asLecturer, 2, "the actor gets every call they made")
	s.True(asLecturer[0].MadeBySubject)
	s.Equal(privacyDTO.ContentWithheld, asLecturer[0].ContentState, "the actor's prompt is about students")

	nobody, err := s.service.exportCalls(s.ctx, "", []uuid.UUID{uuid.New()})
	s.Require().NoError(err)
	s.Empty(nobody)
}

func (s *PrivacySuite) TestErasureRestrictsHighRiskAndDeletesTheRest() {
	highRisk := s.completedCall("assessment.action_item_suggestions", s.student, s.other)
	lowRisk := s.completedCall(lowRiskFeature, s.student, s.other)
	unrelated := s.completedCall(lowRiskFeature, s.other)

	s.Require().NoError(s.service.Delete(s.ginContext(), keycloakTokenVerifier.SubjectIdentifiers{
		UserID: uuid.New(), CourseParticipationIDs: []uuid.UUID{s.student},
	}))

	s.Equal(1, s.count("SELECT count(*) FROM ai_call_content_restriction WHERE call_id = $1", highRisk))
	s.Equal(1, s.count("SELECT count(*) FROM ai_call_content WHERE call_id = $1", highRisk), "restricted content is kept")
	s.Zero(s.count("SELECT count(*) FROM ai_call_content WHERE call_id = $1", lowRisk))
	s.Zero(s.count("SELECT count(*) FROM ai_call_subject WHERE call_id = $1 AND course_participation_id = $2", lowRisk, s.student))
	s.Equal(1, s.count("SELECT count(*) FROM ai_call_subject WHERE call_id = $1", lowRisk), "the other subject keeps their link")
	s.Equal(1, s.count("SELECT count(*) FROM ai_call_content WHERE call_id = $1", unrelated))
	s.Equal(3, s.count("SELECT count(*) FROM ai_call"), "the metadata stays as logging evidence")

	exported, err := s.service.exportCalls(s.ctx, "", []uuid.UUID{s.other})
	s.Require().NoError(err)
	s.Require().Len(exported, 3)
	s.Equal(callDTO.ContentRestricted, exported[0].ContentState)
	s.Nil(exported[0].Request, "restricted content is not handed out")
	s.Equal(callDTO.ContentUnavailable, exported[1].ContentState)
	s.Equal(callDTO.ContentAvailable, exported[2].ContentState)

	s.Require().NoError(s.service.Delete(s.ginContext(), keycloakTokenVerifier.SubjectIdentifiers{
		UserID: uuid.New(), CourseParticipationIDs: []uuid.UUID{s.student},
	}), "erasure is idempotent")
}
