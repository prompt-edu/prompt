package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/auth/authDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// GetCourseRoles returns the roles that grant access to the course phase. It lists the Admin role of
// the course's org only for a token issued after the org was created, as the core checks do, since
// an older token holds the roles of a deleted org with the same slug. The mapping therefore depends
// on the caller's token and must not be cached across callers.
func (s *AuthService) GetCourseRoles(ctx context.Context, coursePhaseID uuid.UUID, tokenIssuedAt time.Time) (authDTO.GetCourseRoles, error) {
	// Get course roles
	courseRoles, err := s.queries.GetCoursePhaseAuthRoleMapping(ctx, coursePhaseID)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"error":         err,
		}).Error("Failed to get course roles")
		return authDTO.GetCourseRoles{}, err
	}

	roles := authDTO.GetCourseRoles{
		CourseLecturerRole: courseRoles.LecturerRole,
		CourseEditorRole:   courseRoles.EditorRole,
		CustomRolePrefix:   courseRoles.CustomRolePrefix,
	}
	if courseRoles.OrgSlug.Valid && permissionValidation.TokenPostdatesOrg(tokenIssuedAt, courseRoles.OrgCreatedAt.Time) {
		roles.OrgAdminRoles = []string{permissionValidation.OrgRoleName(courseRoles.OrgSlug.String, permissionValidation.OrgAdmin)}
	}
	return roles, nil
}

func (s *AuthService) GetCoursePhaseParticipation(ctx context.Context, coursePhaseID uuid.UUID, matriculationNumber string, universityLogin string) (authDTO.GetCoursePhaseParticipation, error) {
	// Get course phase participation
	coursePhaseParticipation, err := s.queries.IsStudentInCoursePhase(ctx, db.IsStudentInCoursePhaseParams{
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
