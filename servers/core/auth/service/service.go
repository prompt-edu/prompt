package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation/courseParticipationDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// CourseParticipationProvider lists the course participations of a student.
type CourseParticipationProvider interface {
	GetAllCourseParticipationsForStudent(ctx context.Context, id uuid.UUID) ([]courseParticipationDTO.GetCourseParticipation, error)
}

type AuthService struct {
	queries              db.Queries
	courseParticipations CourseParticipationProvider
}

func NewAuthService(queries db.Queries, courseParticipations CourseParticipationProvider) *AuthService {
	return &AuthService{queries: queries, courseParticipations: courseParticipations}
}
