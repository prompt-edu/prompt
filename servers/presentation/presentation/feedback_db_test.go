package presentation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/presentation/testutils"
)

var (
	individualPhaseID = uuid.MustParse("10000000-0000-0000-0000-000000000001")
	sharedPhaseID     = uuid.MustParse("10000000-0000-0000-0000-000000000002")
	// Has categories but no presentations or feedback, so category mutations are unlocked.
	categoryOnlyPhaseID = uuid.MustParse("10000000-0000-0000-0000-000000000003")
	deliveryCategory    = uuid.MustParse("20000000-0000-0000-0000-000000000001")
	contentCategory     = uuid.MustParse("20000000-0000-0000-0000-000000000002")
	teamworkCategory    = uuid.MustParse("20000000-0000-0000-0000-000000000003")
	adaPresentationID   = uuid.MustParse("40000000-0000-0000-0000-000000000001")
	teamPresentationID  = uuid.MustParse("40000000-0000-0000-0000-000000000003")
	freeSlotID          = uuid.MustParse("30000000-0000-0000-0000-000000000003")
	assignedSlotID      = uuid.MustParse("30000000-0000-0000-0000-000000000001")
)

func instructor(id string) User {
	return User{ID: id, Name: "Instructor " + id, Email: id + "@example.org", Staff: true}
}

type FeedbackDBTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	service *Service
}

func (s *FeedbackDBTestSuite) SetupSuite() {
	s.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(s.ctx, "../database_dumps/presentation_seed.sql")
	require.NoError(s.T(), err)
	s.cleanup = cleanup
	s.service = NewService(
		testDB.Queries, testDB.Conn, testutils.NewFakeStorage(), "http://core.test",
		60, 60, 50*1024*1024, nil,
	)
}

func (s *FeedbackDBTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *FeedbackDBTestSuite) put(
	categoryID uuid.UUID, user User, value string, expectedRevision int64,
) (FeedbackAnswerResponse, error) {
	return s.service.PutFeedbackAnswer(
		s.ctx, individualPhaseID, adaPresentationID, categoryID, user,
		PutAnswerRequest{Value: value, ExpectedRevision: expectedRevision},
	)
}

// The original single-statement upsert could never reach its update branch, so every
// answer was write-once and each later edit returned 409.
func (s *FeedbackDBTestSuite) TestAnswerCanBeEditedRepeatedly() {
	user := instructor("editable")

	first, err := s.put(deliveryCategory, user, "first", 0)
	require.NoError(s.T(), err)
	assert.EqualValues(s.T(), 1, first.Revision)
	assert.Equal(s.T(), "first", first.Value)

	second, err := s.put(deliveryCategory, user, "second", first.Revision)
	require.NoError(s.T(), err)
	assert.EqualValues(s.T(), 2, second.Revision)
	assert.Equal(s.T(), "second", second.Value)

	third, err := s.put(deliveryCategory, user, "third", second.Revision)
	require.NoError(s.T(), err)
	assert.EqualValues(s.T(), 3, third.Revision)
	assert.Equal(s.T(), "third", third.Value)
}

func (s *FeedbackDBTestSuite) TestStaleRevisionConflictsAndReportsCurrentAnswer() {
	user := instructor("stale")

	created, err := s.put(contentCategory, user, "original", 0)
	require.NoError(s.T(), err)
	_, err = s.put(contentCategory, user, "updated by someone else", created.Revision)
	require.NoError(s.T(), err)

	_, err = s.put(contentCategory, user, "based on stale read", created.Revision)

	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 409, apiErr.Status)
	assert.Equal(s.T(), "feedback_conflict", apiErr.Code)
	// The banner needs the winning value to show what it would have overwritten.
	current, ok := apiErr.Details.(FeedbackAnswerResponse)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "updated by someone else", current.Value)
}

// Two saves that both believe the answer is new: one insert wins, the other conflicts
// rather than silently overwriting. Same instructor, because in independent mode each
// instructor owns a separate form and two users would not contend at all.
func (s *FeedbackDBTestSuite) TestConcurrentCreateYieldsOneWinner() {
	user := instructor("race")

	_, err := s.put(contentCategory, user, "created first", 0)
	require.NoError(s.T(), err)

	_, err = s.put(contentCategory, user, "created second", 0)
	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), "feedback_conflict", apiErr.Code)
}

func (s *FeedbackDBTestSuite) TestDuplicateCategoryPositionConflicts() {
	_, err := s.service.CreateCategory(s.ctx, categoryOnlyPhaseID, CategoryRequest{
		Name:     "Clashing position",
		Position: 0, // Already taken by "Delivery" in the seed.
	}, false)

	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 409, apiErr.Status)
	assert.Equal(s.T(), "category_position_taken", apiErr.Code)
}

func (s *FeedbackDBTestSuite) TestDuplicateCategoryNameConflicts() {
	_, err := s.service.CreateCategory(s.ctx, categoryOnlyPhaseID, CategoryRequest{
		Name:     "Delivery", // Already taken by the seed.
		Position: 99,
	}, false)

	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 409, apiErr.Status)
	assert.Equal(s.T(), "category_name_taken", apiErr.Code)
}

// A missing slot used to be reported as 409 "unassign the presentation first", which
// describes a slot the caller cannot even see.
func (s *FeedbackDBTestSuite) TestDeleteSlotDistinguishesMissingFromAssigned() {
	err := s.service.DeleteSlot(s.ctx, individualPhaseID, uuid.New())
	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 404, apiErr.Status)
	assert.Equal(s.T(), "slot_not_found", apiErr.Code)

	err = s.service.DeleteSlot(s.ctx, individualPhaseID, assignedSlotID)
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 409, apiErr.Status)
	assert.Equal(s.T(), "slot_assigned", apiErr.Code)

	require.NoError(s.T(), s.service.DeleteSlot(s.ctx, individualPhaseID, freeSlotID))
}

// Release must not be able to land between a writer's check and its write.
func (s *FeedbackDBTestSuite) TestSubmitCannotLandAfterRelease() {
	phaseID := individualPhaseID
	presentationID := uuid.MustParse("40000000-0000-0000-0000-000000000002")
	user := instructor("releaser")
	user.CanRelease = true

	_, err := s.service.PutFeedbackAnswer(
		s.ctx, phaseID, presentationID, deliveryCategory, user,
		PutAnswerRequest{Value: "ready to submit", ExpectedRevision: 0},
	)
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.service.SubmitFeedback(s.ctx, phaseID, presentationID, user))
	require.NoError(s.T(), s.service.ReleaseFeedback(s.ctx, phaseID, presentationID, user, "Final"))

	// Everything that mutates feedback must now refuse.
	_, err = s.service.PutFeedbackAnswer(
		s.ctx, phaseID, presentationID, deliveryCategory, user,
		PutAnswerRequest{Value: "sneaking in late", ExpectedRevision: 1},
	)
	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), "feedback_released", apiErr.Code)

	err = s.service.ReopenFeedback(s.ctx, phaseID, presentationID, user)
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), "feedback_released", apiErr.Code)

	err = s.service.DeleteDraft(s.ctx, phaseID, presentationID, user)
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), "feedback_released", apiErr.Code)
}

// The delete is scoped to the shared form in shared mode, so an instructor who means to
// discard their own writing would take the whole group's evaluation with them.
func (s *FeedbackDBTestSuite) TestSharedEvaluationDeleteIsLecturerOnly() {
	author := instructor("shared-author")
	_, err := s.service.PutFeedbackAnswer(
		s.ctx, sharedPhaseID, teamPresentationID, teamworkCategory, author,
		PutAnswerRequest{Value: "written by the group", ExpectedRevision: 0},
	)
	require.NoError(s.T(), err)

	err = s.service.DeleteDraft(s.ctx, sharedPhaseID, teamPresentationID, instructor("shared-peer"))
	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 403, apiErr.Status)
	assert.Equal(s.T(), "shared_feedback_forbidden", apiErr.Code)

	// The form the peer was not allowed to delete is still there.
	_, err = s.service.queries.GetFeedbackFormByScope(s.ctx, db.GetFeedbackFormByScopeParams{
		PresentationID: teamPresentationID,
		ScopeKey:       "shared",
	})
	require.NoError(s.T(), err)

	lecturer := instructor("shared-lecturer")
	lecturer.CanRelease = true
	require.NoError(s.T(), s.service.DeleteDraft(s.ctx, sharedPhaseID, teamPresentationID, lecturer))
}

// The 404 used to be raised after the transaction had committed, so deleting a category
// that another lecturer had already removed destroyed every evaluation in the phase.
func (s *FeedbackDBTestSuite) TestDeletingUnknownCategoryKeepsExistingFeedback() {
	_, err := s.put(deliveryCategory, instructor("category-delete"), "must survive", 0)
	require.NoError(s.T(), err)

	err = s.service.DeleteCategory(s.ctx, individualPhaseID, uuid.New(), true)
	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 404, apiErr.Status)
	assert.Equal(s.T(), "category_not_found", apiErr.Code)

	remaining, err := s.service.queries.CountFeedbackFormsByPhase(s.ctx, individualPhaseID)
	require.NoError(s.T(), err)
	assert.Positive(s.T(), remaining)
}

// The listing pages read their counts from one aggregate query rather than two extra round
// trips per presentation, so both paths have to keep reporting the same numbers.
func (s *FeedbackDBTestSuite) TestListedPresentationCountsMatchTheSingleReads() {
	_, err := s.put(deliveryCategory, instructor("counted"), "counted", 0)
	require.NoError(s.T(), err)

	listed, err := s.service.ListPresentations(s.ctx, individualPhaseID)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), listed)

	counted := false
	for _, presentation := range listed {
		dependencies, err := s.service.queries.CountPresentationDependencies(s.ctx, presentation.ID)
		require.NoError(s.T(), err)
		submitted, err := s.service.queries.CountSubmittedFeedbackForms(s.ctx, presentation.ID)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), dependencies.MaterialCount, presentation.MaterialCount)
		assert.Equal(s.T(), dependencies.FeedbackCount, presentation.FeedbackCount)
		assert.Equal(s.T(), submitted, presentation.SubmittedFeedbackCount)
		counted = counted || presentation.FeedbackCount > 0
	}
	assert.True(s.T(), counted, "the comparison is meaningless if every count is zero")
}

func (s *FeedbackDBTestSuite) TestGetConfigReadsWithoutCreatingRow() {
	unconfigured := uuid.New()

	config, err := s.service.GetConfig(s.ctx, unconfigured)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), targetModeIndividual, config.TargetMode)
	assert.Equal(s.T(), feedbackIndependent, config.FeedbackMode)
	assert.EqualValues(s.T(), 50*1024*1024, config.MaxUploadBytes)

	// Reading must not have written a config row.
	_, err = s.service.queries.GetCoursePhaseConfig(s.ctx, unconfigured)
	assert.Error(s.T(), err)
}

func TestFeedbackDBTestSuite(t *testing.T) {
	suite.Run(t, new(FeedbackDBTestSuite))
}

func TestMaterialReclaimDeadlineOutlivesPresign(t *testing.T) {
	// The presigned URL only has to cover the PUT; the row must survive long enough for a
	// slow upload to be completed afterwards.
	assert.Greater(t, materialReclaimAfter, time.Hour)
}
