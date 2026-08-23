package presentation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/prompt-edu/prompt/servers/presentation/testutils"
)

type SlotDBTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	service *Service
}

func (s *SlotDBTestSuite) SetupSuite() {
	s.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(s.ctx, "../database_dumps/presentation_seed.sql")
	require.NoError(s.T(), err)
	s.cleanup = cleanup
	s.service = NewService(
		testDB.Queries, testDB.Conn, testutils.NewFakeStorage(), "http://core.test",
		60, 60, 50*1024*1024, nil,
	)
}

func (s *SlotDBTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func series(start time.Time, count int, duration, breakBetween time.Duration) []SlotRequest {
	requests := make([]SlotRequest, 0, count)
	for index := range count {
		slotStart := start.Add(time.Duration(index) * (duration + breakBetween))
		requests = append(requests, SlotRequest{
			StartTime: slotStart,
			EndTime:   slotStart.Add(duration),
			Location:  " Room 3 ",
		})
	}
	return requests
}

func (s *SlotDBTestSuite) TestCreateSlotsCreatesTheWholeSeries() {
	start := time.Date(2998, 5, 4, 9, 0, 0, 0, time.UTC)

	created, err := s.service.CreateSlots(s.ctx, individualPhaseID,
		series(start, 3, 20*time.Minute, 5*time.Minute))
	require.NoError(s.T(), err)
	require.Len(s.T(), created, 3)

	assert.Equal(s.T(), "Room 3", created[0].Location, "the location is stored trimmed")
	assert.Equal(s.T(), start.Add(20*time.Minute).UTC(), created[0].EndTime.UTC())
	assert.Equal(s.T(), start.Add(25*time.Minute).UTC(), created[1].StartTime.UTC(),
		"the break has to sit between consecutive slots")

	slots, err := s.service.ListSlots(s.ctx, individualPhaseID)
	require.NoError(s.T(), err)
	for _, slot := range created {
		assert.True(s.T(), func() bool {
			for _, stored := range slots {
				if stored.ID == slot.ID {
					return true
				}
			}
			return false
		}(), "every created slot has to be listed")
	}
}

// One invalid entry must not leave half a series behind for the lecturer to clean up.
func (s *SlotDBTestSuite) TestCreateSlotsRejectsTheWholeBatchOnInvalidSlot() {
	before, err := s.service.ListSlots(s.ctx, individualPhaseID)
	require.NoError(s.T(), err)

	start := time.Date(2997, 5, 4, 9, 0, 0, 0, time.UTC)
	requests := series(start, 2, 20*time.Minute, 0)
	requests = append(requests, SlotRequest{StartTime: start, EndTime: start})

	_, err = s.service.CreateSlots(s.ctx, individualPhaseID, requests)
	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 400, apiErr.Status)
	assert.Equal(s.T(), "invalid_slot", apiErr.Code)

	after, err := s.service.ListSlots(s.ctx, individualPhaseID)
	require.NoError(s.T(), err)
	assert.Len(s.T(), after, len(before), "no slot of a rejected batch may be created")
}

func (s *SlotDBTestSuite) TestCreateSlotsRejectsOversizedBatch() {
	start := time.Date(2996, 5, 4, 9, 0, 0, 0, time.UTC)

	_, err := s.service.CreateSlots(s.ctx, individualPhaseID,
		series(start, MaxBatchSlots+1, time.Minute, 0))

	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), "too_many_slots", apiErr.Code)
}

func TestSlotDBTestSuite(t *testing.T) {
	suite.Run(t, new(SlotDBTestSuite))
}
