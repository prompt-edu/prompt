package courseParticipationDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type CreateCourseParticipation struct {
	CourseID  uuid.UUID `json:"courseID"`
	StudentID uuid.UUID `json:"studentID"`
}

func (c CreateCourseParticipation) GetDBModel() db.CreateCourseParticipationParams {
	return db.CreateCourseParticipationParams{
		CourseID:  c.CourseID,
		StudentID: c.StudentID,
	}
}
