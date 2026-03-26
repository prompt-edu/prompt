package instructorNote

import (
	"context"
	"fmt"
	"time"

	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/instructorNote/instructorNoteDTO"
	promptSDK "github.com/prompt-edu/prompt-sdk"
)

type InstructorNoteService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var InstructorNoteServiceSingleton *InstructorNoteService

func GetStudentNotes(ctx context.Context) ([]instructorNoteDTO.InstructorNote, error) {
  instructorNotes, err := InstructorNoteServiceSingleton.queries.GetAllStudentNotes(ctx)
  if err != nil {
    return nil, err
  }
  return instructorNoteDTO.InstructorNotesFromDBModelToDTO(instructorNotes)
}

func GetStudentNotesByID(ctx context.Context, id uuid.UUID) ([]instructorNoteDTO.InstructorNote, error) {
  instructorNotes, err := InstructorNoteServiceSingleton.queries.GetStudentNotesForStudent(ctx, id)
  if err != nil {
    return nil, err
  }
  return instructorNoteDTO.InstructorNotesFromDBModelToDTO(instructorNotes)
}

func GetStudentNotesByIDWithoutAuthor(ctx context.Context, id uuid.UUID) ([]instructorNoteDTO.InstructorNote, error) {
  notes, err := GetStudentNotesByID(ctx, id)
  if err != nil {
    return nil, err
  }
  for i := range notes {
    notes[i].Author = uuid.Nil
    notes[i].AuthorName = ""
    notes[i].AuthorEmail = ""
  }
  return notes, nil
}

func GetStudentNotesForAuthor(ctx context.Context, authorID uuid.UUID) ([]instructorNoteDTO.InstructorNote, error) {
  instructorNotes, err := InstructorNoteServiceSingleton.queries.GetStudentNotesForAuthor(ctx, authorID)
  if err != nil {
    return nil, err
  }
  return instructorNoteDTO.InstructorNotesFromDBModelToDTO(instructorNotes)
}

func GetStudentNotesForAuthorWithoutStudent(ctx context.Context, authorID uuid.UUID) ([]instructorNoteDTO.InstructorNote, error) {
  notes, err := GetStudentNotesForAuthor(ctx, authorID)
  if err != nil {
    return nil, err
  }
  for i := range notes {
    notes[i].ForStudent = uuid.Nil
  }
  return notes, nil
}


func GetAllTags(ctx context.Context) ([]instructorNoteDTO.NoteTag, error) {
  tags, err := InstructorNoteServiceSingleton.queries.GetAllTags(ctx)
  if err != nil {
    return nil, err
  }
  return instructorNoteDTO.NoteTagsFromDBModels(tags), nil
}

func GetSingleNoteByID(ctx context.Context, id uuid.UUID) (db.Note, error) {
  note, err := InstructorNoteServiceSingleton.queries.GetSingleStudentNoteByID(ctx, id)
  if err != nil {
    return db.Note{}, err
  }
  return note, nil
}

func NewStudentNote(ctx context.Context, studentID uuid.UUID, params instructorNoteDTO.CreateInstructorNote, signedInUserUUID uuid.UUID, authorName string, authorEmail string) (instructorNoteDTO.InstructorNote, error) {
  tx, err := InstructorNoteServiceSingleton.conn.Begin(ctx)
  if err != nil {
    return instructorNoteDTO.InstructorNote{}, err
  }
  defer promptSDK.DeferDBRollback(tx, ctx)
  qtx := InstructorNoteServiceSingleton.queries.WithTx(tx)

  versionNumber := 0
  var noteID uuid.UUID
  rightNow := pgtype.Timestamptz{Time: time.Now(), Valid: true}

  if params.New {
    newNoteId, err := uuid.NewRandom()
    if err != nil {
      return instructorNoteDTO.InstructorNote{}, err
    }
    _, err = qtx.CreateNote(ctx, db.CreateNoteParams{
      ID:          newNoteId,
      ForStudent:  studentID,
      Author:      signedInUserUUID,
      AuthorName:  authorName,
      AuthorEmail: authorEmail,
      DateCreated: rightNow,
      DateDeleted: pgtype.Timestamptz{},
      DeletedBy:   pgtype.UUID{},
    })
    if err != nil {
      return instructorNoteDTO.InstructorNote{}, err
    }
    noteID = newNoteId
  } else {
    noteID = params.ForNote
    latestVersionNumber, err := qtx.GetLatestNoteVersionForNoteId(ctx, params.ForNote)
    if err != nil {
      return instructorNoteDTO.InstructorNote{}, err
    }
    versionNumber = int(latestVersionNumber) + 1
  }

  versionID, err := uuid.NewRandom()
  if err != nil {
    return instructorNoteDTO.InstructorNote{}, err
  }

  _, err = qtx.CreateNoteVersion(ctx, db.CreateNoteVersionParams{
    ID:            versionID,
    Content:       params.Content,
    DateCreated:   rightNow,
    VersionNumber: int32(versionNumber),
    ForNote:       noteID,
  })
  if err != nil {
    return instructorNoteDTO.InstructorNote{}, err
  }

  if !params.New {
    if err := qtx.RemoveAllTagsFromNote(ctx, noteID); err != nil {
      return instructorNoteDTO.InstructorNote{}, err
    }
  }
  for _, tagID := range params.Tags {
    if _, err := qtx.GetTagByID(ctx, tagID); err != nil {
      if errors.Is(err, pgx.ErrNoRows) {
        return instructorNoteDTO.InstructorNote{}, fmt.Errorf("tag %s does not exist", tagID)
      }
      return instructorNoteDTO.InstructorNote{}, err
    }
    if err := qtx.AddTagToNote(ctx, db.AddTagToNoteParams{
      NoteID: noteID,
      TagID:  tagID,
    }); err != nil {
      return instructorNoteDTO.InstructorNote{}, err
    }
  }

  if err := tx.Commit(ctx); err != nil {
    return instructorNoteDTO.InstructorNote{}, fmt.Errorf("failed to commit transaction: %w", err)
  }

  createdNote, err := InstructorNoteServiceSingleton.queries.GetSingleNoteWithVersionsByID(ctx, noteID)
  if err != nil {
    return instructorNoteDTO.InstructorNote{}, err
  }
  return instructorNoteDTO.GetInstructorNoteDTOFromDBModel(createdNote)
}

func CreateNoteTag(ctx context.Context, tag instructorNoteDTO.CreateNoteTag) (instructorNoteDTO.NoteTag, error) {
  id, err := uuid.NewRandom()
  if err != nil {
    return instructorNoteDTO.NoteTag{}, err
  }
  result, err := InstructorNoteServiceSingleton.queries.CreateTag(ctx, db.CreateTagParams{
    ID:    id,
    Name:  tag.Name,
    Color: db.NoteTagColor(tag.Color),
  })
  if err != nil {
    return instructorNoteDTO.NoteTag{}, err
  }
  return instructorNoteDTO.NoteTagFromDBModel(result), nil
}

func UpdateNoteTag(ctx context.Context, id uuid.UUID, tag instructorNoteDTO.UpdateNoteTag) (instructorNoteDTO.NoteTag, error) {
  result, err := InstructorNoteServiceSingleton.queries.UpdateTag(ctx, db.UpdateTagParams{
    ID:    id,
    Name:  tag.Name,
    Color: db.NoteTagColor(tag.Color),
  })
  if err != nil {
    return instructorNoteDTO.NoteTag{}, err
  }
  return instructorNoteDTO.NoteTagFromDBModel(result), nil
}

func DeleteNoteTag(ctx context.Context, id uuid.UUID) error {
  return InstructorNoteServiceSingleton.queries.DeleteTag(ctx, id)
}

func DeleteInstructorNote(ctx context.Context, noteID uuid.UUID, authorID uuid.UUID) (instructorNoteDTO.InstructorNote, error) {
  _, err := InstructorNoteServiceSingleton.queries.DeleteNote(ctx, db.DeleteNoteParams{
    ID: noteID,
    DeletedBy: pgtype.UUID{ Bytes: authorID, Valid: true},
  })
  if err != nil {
    return instructorNoteDTO.InstructorNote{}, err
  }

  // Fetch the deleted note with versions to return to client
  deletedNoteWithVersions, err := InstructorNoteServiceSingleton.queries.GetSingleNoteWithVersionsByID(ctx, noteID)
  if err != nil {
    return instructorNoteDTO.InstructorNote{}, err
  }

  // Convert to DTO
  return instructorNoteDTO.GetInstructorNoteDTOFromDBModel(deletedNoteWithVersions)
}
