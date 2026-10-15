package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
)

// collectProfilePictureFileIDs finds the subject's profile pictures: the one owned by their
// account, and any matched to their student record through the university login. The second
// case covers admin-initiated requests, which know the student but not the account.
func (s *PrivacyService) collectProfilePictureFileIDs(ctx context.Context, subject sdk.SubjectIdentifiers) ([]uuid.UUID, error) {
	seen := map[uuid.UUID]bool{}
	fileIDs := []uuid.UUID{}
	add := func(fileID uuid.UUID) {
		if !seen[fileID] {
			seen[fileID] = true
			fileIDs = append(fileIDs, fileID)
		}
	}

	if subject.UserID != uuid.Nil {
		picture, err := s.queries.GetProfilePictureByUserID(ctx, subject.UserID)
		if err == nil {
			add(picture.FileID)
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get profile picture of user: %w", err)
		}
	}

	if subject.StudentID != uuid.Nil {
		studentFileIDs, err := s.queries.GetProfilePictureFileIDsForStudent(ctx, subject.StudentID)
		if err != nil {
			return nil, fmt.Errorf("get profile pictures of student: %w", err)
		}
		for _, fileID := range studentFileIDs {
			add(fileID)
		}
	}

	return fileIDs, nil
}
