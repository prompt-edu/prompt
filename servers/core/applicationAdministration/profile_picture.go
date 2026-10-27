package applicationAdministration

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

const profilePictureRequired = "required"

// ErrProfilePictureRequired marks an application submitted without the picture the phase requires.
var ErrProfilePictureRequired = errors.New("this application requires a profile picture")

// ValidateProfilePictureRequirement rejects a logged-in applicant without a profile picture when
// the application requires one. The form checks this as well; this guards direct API calls.
func (s *ApplicationService) ValidateProfilePictureRequirement(ctx context.Context, coursePhaseID uuid.UUID, userID uuid.UUID) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	applicationPhase, err := s.queries.GetOpenApplicationPhase(ctxWithTimeout, coursePhaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		// A closed or unknown application is rejected by the regular validation
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to load the application phase: %w", err)
	}
	if applicationPhase.ProfilePictureRequirement != profilePictureRequired {
		return nil
	}

	_, err = s.queries.GetProfilePictureByUserID(ctxWithTimeout, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrProfilePictureRequired
	}
	if err != nil {
		return fmt.Errorf("failed to load the profile picture: %w", err)
	}
	return nil
}
