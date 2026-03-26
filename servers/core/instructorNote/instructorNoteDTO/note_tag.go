package instructorNoteDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type NoteTag struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

type CreateNoteTag struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type UpdateNoteTag struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func NoteTagFromDBModel(model db.NoteTag) NoteTag {
	return NoteTag{
		ID:    model.ID,
		Name:  model.Name,
		Color: string(model.Color),
	}
}

func NoteTagsFromDBModels(models []db.NoteTag) []NoteTag {
	tags := make([]NoteTag, 0, len(models))
	for _, m := range models {
		tags = append(tags, NoteTagFromDBModel(m))
	}
	return tags
}
