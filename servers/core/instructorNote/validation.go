package instructorNote

import (
	"context"
	"errors"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/instructorNote/instructorNoteDTO"
)

func ValidateCreateNote(createNoteRequest instructorNoteDTO.CreateInstructorNote) error {
	if !createNoteRequest.New && createNoteRequest.ForNote == uuid.Nil {
		return errors.New("for editing existing notes, a note id must be provided")
	}
	return nil
}

// VerifyNoteOwnership checks that the given user is the author of the note.
// Returns the note on success so callers can perform additional checks without re-fetching.
func VerifyNoteOwnership(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (db.Note, error) {
	note, err := GetSingleNoteByID(ctx, noteID)
	if err != nil {
		return db.Note{}, err
	}
	if note.Author != userID {
		return db.Note{}, errors.New("the user that performed the request is not the author of the InstructorNote")
	}
	return note, nil
}

func ValidateReferencedNote(createRequest instructorNoteDTO.CreateInstructorNote, ctx context.Context, SignedInUserID uuid.UUID) error {
	if !createRequest.New {
		note, err := VerifyNoteOwnership(ctx, createRequest.ForNote, SignedInUserID)
		if err != nil {
			return err
		}
		if note.DateDeleted.Valid {
			return errors.New("cannot edit a deleted note")
		}
	}
	return nil
}
