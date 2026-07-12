package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type TeamsPrivacyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var TeamsPrivacyServiceSingleton *TeamsPrivacyService

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Team Allocation", "team_allocation.json", func() (any, error) {
		return TeamsPrivacyServiceSingleton.queries.GetAllocationByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Team Preferences", "team_preferences.json", func() (any, error) {
		return TeamsPrivacyServiceSingleton.queries.GetStudentTeamPreferenceResponseByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Skill Responses", "skill_responses.json", func() (any, error) {
		return TeamsPrivacyServiceSingleton.queries.GetStudentSkillResponseByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Tutor Assignment", "tutor.json", func() (any, error) {
		return TeamsPrivacyServiceSingleton.queries.GetTutorByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	return nil
}

func PrivacyDataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	tx, err := TeamsPrivacyServiceSingleton.conn.Begin(c)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, c)
	qtx := TeamsPrivacyServiceSingleton.queries.WithTx(tx)

	if err := qtx.DeleteAllocationsByCourseParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	if err := qtx.DeleteStudentTeamPreferenceResponseByCourseParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	if err := qtx.DeleteStudentSkillResponseByCourseParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	if err := qtx.DeleteTutorByCourseParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}

	return tx.Commit(c)
}
