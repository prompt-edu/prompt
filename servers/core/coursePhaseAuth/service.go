package coursePhaseAuth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseAuth/coursePhaseAuthDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type CoursePhaseAuthService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CoursePhaseAuthServiceSingleton *CoursePhaseAuthService

func GetCourseRoles(ctx context.Context, coursePhaseID uuid.UUID) (coursePhaseAuthDTO.GetCourseRoles, error) {
	// Get course roles
	courseRoles, err := CoursePhaseAuthServiceSingleton.queries.GetCoursePhaseAuthRoleMapping(ctx, coursePhaseID)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"error":         err,
		}).Error("Failed to get course roles")
		return coursePhaseAuthDTO.GetCourseRoles{}, err
	}

	return coursePhaseAuthDTO.GetCourseRoles{
		CourseLecturerRole: courseRoles.LecturerRole,
		CourseEditorRole:   courseRoles.EditorRole,
		CustomRolePrefix:   courseRoles.CustomRolePrefix,
	}, nil
}

func GetCoursePhaseParticipation(ctx context.Context, coursePhaseID uuid.UUID, matriculationNumber string, universityLogin string) (coursePhaseAuthDTO.GetCoursePhaseParticipation, error) {
	// Get course phase participation
	coursePhaseParticipation, err := CoursePhaseAuthServiceSingleton.queries.IsStudentInCoursePhase(ctx, db.IsStudentInCoursePhaseParams{
		CoursePhaseID:       coursePhaseID,
		MatriculationNumber: matriculationNumber,
		UniversityLogin:     universityLogin,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID":       coursePhaseID,
			"matriculationNumber": matriculationNumber,
			"universityLogin":     universityLogin,
			"error":               err,
		}).Error("Failed to get course phase participation")
		return coursePhaseAuthDTO.GetCoursePhaseParticipation{}, err
	}

	return coursePhaseAuthDTO.GetCoursePhaseParticipation{
		IsStudentOfCoursePhase: coursePhaseParticipation.IsInPhase,
		CourseParticipationID:  coursePhaseParticipation.CourseParticipationID,
	}, nil
}
