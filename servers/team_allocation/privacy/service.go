package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func NewTeamsPrivacyService(queries db.Queries, conn *pgxpool.Pool) *TeamsPrivacyService {
	return &TeamsPrivacyService{
		queries: queries,
		conn:    conn,
	}
}

func (s *TeamsPrivacyService) DataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Team Allocation", "team_allocation.json", func() (any, error) {
		return s.queries.GetAllocationByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Team Preferences", "team_preferences.json", func() (any, error) {
		return s.queries.GetStudentTeamPreferenceResponseByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Skill Responses", "skill_responses.json", func() (any, error) {
		return s.queries.GetStudentSkillResponseByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Tutor Assignment", "tutor.json", func() (any, error) {
		return s.queries.GetTutorByCourseParticipationIDs(c, tutorIDs(subject))
	})
	return nil
}

func (s *TeamsPrivacyService) DataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	tx, err := s.conn.Begin(c)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, c)
	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteAllocationsByCourseParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	if err := qtx.DeleteStudentTeamPreferenceResponseByCourseParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	if err := qtx.DeleteStudentSkillResponseByCourseParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	if err := qtx.DeleteTutorByCourseParticipationIDs(c, tutorIDs(subject)); err != nil {
		return err
	}

	return tx.Commit(c)
}

// tutorIDs lists the ids a tutor row of the subject can carry: a coach is stored under a course
// participation, a course editor (a project lead, for example) under their Keycloak user id.
func tutorIDs(subject sdkAuth.SubjectIdentifiers) []uuid.UUID {
	ids := append([]uuid.UUID{}, subject.CourseParticipationIDs...)
	if subject.UserID != uuid.Nil {
		ids = append(ids, subject.UserID)
	}
	return ids
}
