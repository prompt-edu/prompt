package authDTO

import "github.com/google/uuid"

type GetCoursePhaseParticipation struct {
	IsStudentOfCoursePhase bool      `json:"isStudentOfCoursePhase"`
	CourseParticipationID  uuid.UUID `json:"courseParticipationID"`
}
