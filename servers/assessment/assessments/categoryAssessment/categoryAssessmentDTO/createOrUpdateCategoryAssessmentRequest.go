package categoryAssessmentDTO

import (
	"github.com/google/uuid"
)

type CreateOrUpdateCategoryAssessmentRequest struct {
	CategoryID            uuid.UUID `json:"categoryID"`
	CoursePhaseID         uuid.UUID `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Comment               string    `json:"comment"`
	Author                string    `json:"author"`
	AuthorID              string    `json:"authorID"`
}
