package instructorNoteDTO

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type NoteVersion struct {
	ID            uuid.UUID `json:"id"`
	Content       string    `json:"content"`
	DateCreated   time.Time `json:"dateCreated"`
	VersionNumber int32     `json:"versionNumber"`
}

type InstructorNote struct {
	ID          uuid.UUID      `json:"id"`
	Author      uuid.UUID      `json:"author"`
  AuthorName  string         `json:"authorName"`
  AuthorEmail string         `json:"authorEmail"`
	ForStudent  uuid.UUID      `json:"forStudent"`
	DateCreated time.Time      `json:"dateCreated"`
	DateDeleted *time.Time     `json:"dateDeleted,omitempty"`
	DeletedBy   *uuid.UUID     `json:"deletedBy,omitempty"`
	Versions    []NoteVersion  `json:"versions"`
	Tags        []NoteTag      `json:"tags"`
}

func GetInstructorNoteDTOFromDBModel(model db.NoteWithVersion) (InstructorNote, error) {
	// Parse versions JSONB
	var versions []NoteVersion
	if err := json.Unmarshal(model.Versions, &versions); err != nil {
		return InstructorNote{}, err
	}

	// Parse tags JSONB
	var tags []NoteTag
	if err := json.Unmarshal(model.Tags, &tags); err != nil {
		return InstructorNote{}, err
	}

	// date_deleted (pgtype.Date)
	var dateDeleted *time.Time
	if model.DateDeleted.Valid {
		t := model.DateDeleted.Time
		dateDeleted = &t
	}

	// deleted_by (pgtype.UUID)
	var deletedBy *uuid.UUID
	if model.DeletedBy.Valid {
		u, err := uuid.FromBytes(model.DeletedBy.Bytes[:])
		if err != nil {
			return InstructorNote{}, err
		}
		deletedBy = &u
	}

	return InstructorNote{
		ID:          model.ID,
		Author:      model.Author,
    AuthorName:  model.AuthorName,
    AuthorEmail: model.AuthorEmail,
		ForStudent:  model.ForStudent,
		DateCreated: model.DateCreated.Time,
		DateDeleted: dateDeleted,
		DeletedBy:   deletedBy,
		Versions:    versions,
		Tags:        tags,
	}, nil
}


func InstructorNotesFromDBModelToDTO(notes []db.NoteWithVersion) ([]InstructorNote, error) {
	dtoInstructorNotes := make([]InstructorNote, 0, len(notes))
	for _, iN := range notes {
    dtoInstructorNote, err := GetInstructorNoteDTOFromDBModel(iN)
    if err != nil {
      return nil, err
    }
		dtoInstructorNotes = append(dtoInstructorNotes, dtoInstructorNote)
	}
  return dtoInstructorNotes, nil
}
