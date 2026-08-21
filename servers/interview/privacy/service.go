package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

type PrivacyService struct {
	Queries db.Queries
	conn    *pgxpool.Pool
}

var PrivacyServiceSingleton *PrivacyService

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Interview Assignments", "interview_assignments.json", func() (any, error) {
		return PrivacyServiceSingleton.Queries.GetInterviewAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Interview Reviews", "interview_reviews.json", func() (any, error) {
		return PrivacyServiceSingleton.Queries.GetInterviewReviewsByParticipationIDs(c, subject.CourseParticipationIDs)
	})
	return nil
}

func PrivacyDataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	tx, err := PrivacyServiceSingleton.conn.Begin(c)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, c)
	qtx := PrivacyServiceSingleton.Queries.WithTx(tx)

	if err := qtx.DeleteInterviewAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	if err := qtx.DeleteInterviewReviewsByParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	return tx.Commit(c)
}
