package interviewReview

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	"github.com/prompt-edu/prompt/servers/interview/interviewReview/interviewReviewDTO"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InterviewReviewServiceTestSuite struct {
	suite.Suite
	ctx           context.Context
	testDB        *sdkTestUtils.TestDB[*db.Queries]
	cleanup       func()
	activePhaseID uuid.UUID
}

func (suite *InterviewReviewServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup
	suite.activePhaseID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

	InterviewReviewServiceSingleton = &InterviewReviewService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
}

func (suite *InterviewReviewServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *InterviewReviewServiceTestSuite) TestUpsertCreatesAndUpdatesReview() {
	participationID := uuid.New()
	score := int32(2)

	created, err := UpsertInterviewReview(suite.ctx, suite.activePhaseID, participationID, interviewReviewDTO.UpdateInterviewReviewRequest{
		Score:       &score,
		Interviewer: "Ada Lovelace",
		InterviewAnswers: []interviewReviewDTO.InterviewAnswer{
			{QuestionID: 1, Answer: "great"},
		},
	})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), created.Score)
	require.Equal(suite.T(), int32(2), *created.Score)
	require.NotNil(suite.T(), created.ScoreLevel)
	require.Equal(suite.T(), interviewReviewDTO.ScoreLevelGood, *created.ScoreLevel)
	require.Equal(suite.T(), "Ada Lovelace", created.Interviewer)
	require.Len(suite.T(), created.InterviewAnswers, 1)

	// Upsert again to update the same participation.
	updatedScore := int32(5)
	updated, err := UpsertInterviewReview(suite.ctx, suite.activePhaseID, participationID, interviewReviewDTO.UpdateInterviewReviewRequest{
		Score:            &updatedScore,
		Interviewer:      "Alan Turing",
		InterviewAnswers: []interviewReviewDTO.InterviewAnswer{},
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int32(5), *updated.Score)
	require.Equal(suite.T(), interviewReviewDTO.ScoreLevelVeryBad, *updated.ScoreLevel)
	require.Equal(suite.T(), "Alan Turing", updated.Interviewer)
	require.Empty(suite.T(), updated.InterviewAnswers)

	// Exactly one row must exist for the phase/participation combination.
	all, err := GetInterviewReviews(suite.ctx, suite.activePhaseID)
	require.NoError(suite.T(), err)
	count := 0
	for _, r := range all {
		if r.CourseParticipationID == participationID {
			count++
		}
	}
	require.Equal(suite.T(), 1, count)
}

func (suite *InterviewReviewServiceTestSuite) TestScoreAndScoreLevelOutputsOnlyIncludeScored() {
	scoredParticipation := uuid.New()
	unscoredParticipation := uuid.New()
	score := int32(1)

	_, err := UpsertInterviewReview(suite.ctx, suite.activePhaseID, scoredParticipation, interviewReviewDTO.UpdateInterviewReviewRequest{
		Score:       &score,
		Interviewer: "Grace Hopper",
	})
	require.NoError(suite.T(), err)

	_, err = UpsertInterviewReview(suite.ctx, suite.activePhaseID, unscoredParticipation, interviewReviewDTO.UpdateInterviewReviewRequest{
		Score:       nil,
		Interviewer: "Grace Hopper",
	})
	require.NoError(suite.T(), err)

	scores, err := GetScores(suite.ctx, suite.activePhaseID)
	require.NoError(suite.T(), err)
	require.True(suite.T(), containsScore(scores, scoredParticipation))
	require.False(suite.T(), containsScore(scores, unscoredParticipation))

	scoreLevels, err := GetScoreLevels(suite.ctx, suite.activePhaseID)
	require.NoError(suite.T(), err)
	require.True(suite.T(), containsScoreLevel(scoreLevels, scoredParticipation, interviewReviewDTO.ScoreLevelVeryGood))
	require.False(suite.T(), containsScoreLevel(scoreLevels, unscoredParticipation, ""))
}

func containsScore(scores []interviewReviewDTO.ScoreWithParticipation, id uuid.UUID) bool {
	for _, s := range scores {
		if s.CourseParticipationID == id {
			return true
		}
	}
	return false
}

func containsScoreLevel(levels []interviewReviewDTO.ScoreLevelWithParticipation, id uuid.UUID, level interviewReviewDTO.ScoreLevel) bool {
	for _, l := range levels {
		if l.CourseParticipationID == id && (level == "" || l.ScoreLevel == level) {
			return true
		}
	}
	return false
}

func TestInterviewReviewServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InterviewReviewServiceTestSuite))
}

func TestDeriveScoreLevel(t *testing.T) {
	cases := []struct {
		score    int32
		expected interviewReviewDTO.ScoreLevel
		ok       bool
	}{
		{0, "", false},
		{1, interviewReviewDTO.ScoreLevelVeryGood, true},
		{2, interviewReviewDTO.ScoreLevelGood, true},
		{3, interviewReviewDTO.ScoreLevelOk, true},
		{4, interviewReviewDTO.ScoreLevelBad, true},
		{5, interviewReviewDTO.ScoreLevelVeryBad, true},
		{6, "", false},
	}
	for _, tc := range cases {
		level, ok := interviewReviewDTO.DeriveScoreLevel(tc.score)
		require.Equal(t, tc.ok, ok, "score %d", tc.score)
		require.Equal(t, tc.expected, level, "score %d", tc.score)
	}
}
