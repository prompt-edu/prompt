package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/auth/authDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

func GetCourseRoles(ctx context.Context, coursePhaseID uuid.UUID) (authDTO.GetCourseRoles, error) {
	// Get course roles
	courseRoles, err := AuthServiceSingleton.queries.GetCoursePhaseAuthRoleMapping(ctx, coursePhaseID)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"error":         err,
		}).Error("Failed to get course roles")
		return authDTO.GetCourseRoles{}, err
	}

	return authDTO.GetCourseRoles{
		CourseLecturerRole: courseRoles.LecturerRole,
		CourseEditorRole:   courseRoles.EditorRole,
		CustomRolePrefix:   courseRoles.CustomRolePrefix,
	}, nil
}

func GetCoursePhaseParticipation(ctx context.Context, coursePhaseID uuid.UUID, matriculationNumber string, universityLogin string) (authDTO.GetCoursePhaseParticipation, error) {
	// Get course phase participation
	coursePhaseParticipation, err := AuthServiceSingleton.queries.IsStudentInCoursePhase(ctx, db.IsStudentInCoursePhaseParams{
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
		return authDTO.GetCoursePhaseParticipation{}, err
	}

	return authDTO.GetCoursePhaseParticipation{
		IsStudentOfCoursePhase: coursePhaseParticipation.IsInPhase,
		CourseParticipationID:  coursePhaseParticipation.CourseParticipationID,
	}, nil
}
