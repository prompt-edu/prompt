package applicationAdministration

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

const profilePictureRequired = "required"

var ErrProfilePictureRequired = errors.New("this application requires a profile picture")

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

func (s *ApplicationService) SyncProfilePictureUniversityLogin(ctx context.Context, userID uuid.UUID, universityLogin string) error {
	if universityLogin == "" {
		return nil
	}

	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	err := s.queries.SyncProfilePictureUniversityLogin(ctxWithTimeout, db.SyncProfilePictureUniversityLoginParams{
		UserID:          userID,
		UniversityLogin: pgtype.Text{String: universityLogin, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to sync the profile picture login: %w", err)
	}
	return nil
}
