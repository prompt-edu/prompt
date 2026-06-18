package categoryAssessmentDTO

import (
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type CategoryAssessment struct {
	ID                    uuid.UUID `json:"id"`
	CategoryID            uuid.UUID `json:"categoryID"`
	CoursePhaseID         uuid.UUID `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Comment               string    `json:"comment"`
	Author                string    `json:"author"`
	AuthorID              string    `json:"authorID"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

func MapDBCategoryAssessmentToDTO(a db.CategoryAssessment) CategoryAssessment {
	return CategoryAssessment{
		ID:                    a.ID,
		CategoryID:            a.CategoryID,
		CoursePhaseID:         a.CoursePhaseID,
		CourseParticipationID: a.CourseParticipationID,
		Comment:               a.Comment,
		Author:                a.Author,
		AuthorID:              a.AuthorID,
		CreatedAt:             a.CreatedAt.Time,
		UpdatedAt:             a.UpdatedAt.Time,
	}
}

func GetCategoryAssessmentDTOsFromDBModels(dbItems []db.CategoryAssessment) []CategoryAssessment {
	result := make([]CategoryAssessment, 0, len(dbItems))
	for _, a := range dbItems {
		result = append(result, MapDBCategoryAssessmentToDTO(a))
	}
	return result
}
