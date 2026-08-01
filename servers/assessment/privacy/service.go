package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type PrivacyService struct {
	Queries db.Queries
	Conn    *pgxpool.Pool
}

var PrivacyServiceSingleton *PrivacyService

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	q := PrivacyServiceSingleton.Queries
	ids := subject.CourseParticipationIDs

	// The queries backing this export omit every author / third-party identifier
	// at the SQL level, so no redaction is required here.
	exp.AddJSON("Assessments", "student/assessment.json", func() (any, error) {
		return q.GetAllAssessmentsByCourseParticipationIDs(c, ids)
	})
	exp.AddJSON("Assessment Completions", "student/assessment_completion.json", func() (any, error) {
		return q.GetAllAssessmentCompletionsByCourseParticipationIDs(c, ids)
	})
	exp.AddJSON("Evaluations", "student/evaluation.json", func() (any, error) {
		return q.GetAllEvaluationsByCourseParticipationIDs(c, ids)
	})
	exp.AddJSON("Evaluation Completions", "student/evaluation_completion.json", func() (any, error) {
		return q.GetAllEvaluationCompletionsByCourseParticipationIDs(c, ids)
	})
	exp.AddJSON("Action Items", "student/action_item.json", func() (any, error) {
		return q.GetAllActionItemsByCourseParticipationIDs(c, ids)
	})
	exp.AddJSON("Feedback Items", "student/feedback_item.json", func() (any, error) {
		return q.GetAllFeedbackItemsByCourseParticipationIDs(c, ids)
	})

	return nil
}

func PrivacyDataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	ids := subject.CourseParticipationIDs

	tx, err := PrivacyServiceSingleton.Conn.Begin(c)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, c)
	qtx := PrivacyServiceSingleton.Queries.WithTx(tx)

	if err := qtx.DeleteAssessmentsByCourseParticipationIDs(c, ids); err != nil {
		return err
	}
	if err := qtx.DeleteAssessmentCompletionsByCourseParticipationIDs(c, ids); err != nil {
		return err
	}
	if err := qtx.DeleteCategoryAssessmentsByCourseParticipationIDs(c, ids); err != nil {
		return err
	}
	if err := qtx.DeleteEvaluationsByCourseParticipationIDs(c, ids); err != nil {
		return err
	}
	if err := qtx.DeleteEvaluationCompletionsByCourseParticipationIDs(c, ids); err != nil {
		return err
	}
	if err := qtx.DeleteActionItemsByCourseParticipationIDs(c, ids); err != nil {
		return err
	}
	if err := qtx.DeleteFeedbackItemsByCourseParticipationIDs(c, ids); err != nil {
		return err
	}

	return tx.Commit(c)
}
