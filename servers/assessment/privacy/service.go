package privacy

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type PrivacyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewPrivacyService(queries db.Queries, conn *pgxpool.Pool) *PrivacyService {
	return &PrivacyService{
		queries: queries,
		conn:    conn,
	}
}

func (s *PrivacyService) DataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	q := s.queries
	ids := subject.CourseParticipationIDs

	exp.AddJSON("Assessments", "student/assessment.json", func() (any, error) {
		return q.GetAllAssessmentsByCourseParticipationIDs(c, ids)
	})
	exp.AddJSON("Assessment Completions", "student/assessment_completion.json", func() (any, error) {
		return q.GetAllAssessmentCompletionsByCourseParticipationIDs(c, ids)
	})
	exp.AddJSON("Category Assessments", "student/category_assessment.json", func() (any, error) {
		return q.GetAllCategoryAssessmentsByCourseParticipationIDs(c, ids)
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

func (s *PrivacyService) DataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	ids := subject.CourseParticipationIDs

	// needed so when caller closes connection we also stop the procedure
	ctx := c.Request.Context()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin deletion transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteAssessmentsByCourseParticipationIDs(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete assessments: %w", err)
	}
	if err := qtx.DeleteAssessmentCompletionsByCourseParticipationIDs(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete assessment completions: %w", err)
	}
	if err := qtx.DeleteCategoryAssessmentsByCourseParticipationIDs(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete category assessments: %w", err)
	}
	if err := qtx.DeleteEvaluationsByRecipientOrAuthorIDs(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete evaluations: %w", err)
	}
	if err := qtx.DeleteEvaluationCompletionsByRecipientOrAuthorIDs(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete evaluation completions: %w", err)
	}
	if err := qtx.DeleteActionItemsByCourseParticipationIDs(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete action items: %w", err)
	}
	if err := qtx.DeleteFeedbackItemsByRecipientOrAuthorIDs(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete feedback items: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit deletion transaction: %w", err)
	}

	return nil
}
