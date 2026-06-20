package privacy

import (
	"github.com/gin-gonic/gin"
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
		return TeamsPrivacyServiceSingleton.queries.GetAllocationByCourseParticipationID(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Team Preferences", "team_preferences.json", func() (any, error) {
		return TeamsPrivacyServiceSingleton.queries.GetStudentTeamPreferenceResponseByCourseParticipationID(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Skill Responses", "skill_responses.json", func() (any, error) {
		return TeamsPrivacyServiceSingleton.queries.GetStudentSkillResponseByCourseParticipationID(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Tutor Assignment", "tutor.json", func() (any, error) {
		return TeamsPrivacyServiceSingleton.queries.GetTutorsByCourseParticipationID(c, subject.CourseParticipationIDs)
	})
	return nil
}
