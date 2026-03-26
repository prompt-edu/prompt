package instructorNoteDTO

import (
	"github.com/google/uuid"
)

type CreateInstructorNote struct {
  Content string      `json:"content"`
  New     bool        `json:"new"`
  ForNote uuid.UUID   `json:"forNote"`
  Tags    []uuid.UUID `json:"tags,omitempty"`
}

