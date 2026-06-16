package privacy

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type TeamsPrivacyService struct {
	queries db.Queries
}

var TeamsPrivacyServiceSingleton *TeamsPrivacyService

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Team Allocation", "team_allocation.json", func() (any, error) {
		return getTeamForCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Team Preferences", "team_preferences.json", func() (any, error) {
		return getTeamPreferencesForCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Skill Responses", "skill_responses.json", func() (any, error) {
		return getSkillResponsesForCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Tutor Assignment", "tutor.json", func() (any, error) {
		return getTutorForCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	return nil
}

func getTutorForCourseParticipationIDs(ctx context.Context, courseParticipationIDs []uuid.UUID) (any, error) {
	return TeamsPrivacyServiceSingleton.queries.GetTutorByCourseParticipationIDs(ctx, courseParticipationIDs)
}

func getTeamForCourseParticipationIDs(ctx context.Context, courseParticipationIDs []uuid.UUID) (any, error) {
	return TeamsPrivacyServiceSingleton.queries.GetAllocationByCourseParticipationID(ctx, courseParticipationIDs)
}

func getTeamPreferencesForCourseParticipationIDs(ctx context.Context, courseParticipationIDs []uuid.UUID) (any, error) {
	return TeamsPrivacyServiceSingleton.queries.GetStudentTeamPreferenceResponseByCourseParticipationID(ctx, courseParticipationIDs)
}

func getSkillResponsesForCourseParticipationIDs(ctx context.Context, courseParticipationIDs []uuid.UUID) (any, error) {
	return TeamsPrivacyServiceSingleton.queries.GetStudentSkillResponseByCourseParticipationID(ctx, courseParticipationIDs)
}
