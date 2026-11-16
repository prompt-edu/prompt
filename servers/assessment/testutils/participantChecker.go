package testutils

import (
	"context"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
)

// The stubbed core lists every participation in every phase except these two.
var (
	OutsideParticipationID      = uuid.MustParse("6f1c2a4e-3b5d-4e7f-8a9b-0c1d2e3f4a5b")
	UnverifiableParticipationID = uuid.MustParse("7a2d3b5f-4c6e-4f80-9bac-1d2e3f4a5b6c")
)

// CheckTestParticipant stubs coursePhaseConfig.CheckParticipantInPhase: OutsideParticipationID is
// not in the phase, UnverifiableParticipationID fails the lookup, and every other id is listed.
var CheckTestParticipant coursePhaseConfig.ParticipantChecker = func(_ context.Context, _ string, _, courseParticipationID uuid.UUID) error {
	switch courseParticipationID {
	case OutsideParticipationID:
		return coursePhaseConfig.ErrParticipantNotInPhase
	case UnverifiableParticipationID:
		return coursePhaseConfig.ErrParticipantLookup
	}
	return nil
}
