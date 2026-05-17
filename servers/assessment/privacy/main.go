package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type PrivacyService struct {
	Queries db.Queries
	Conn    *pgxpool.Pool
}

var PrivacyServiceSingleton *PrivacyService

func InitPrivacyModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	promptTypes.RegisterPrivacyDataExportEndpoint(routerGroup, PrivacyDataExportHandler, []string{})
	PrivacyServiceSingleton = &PrivacyService{
		Queries: queries,
		Conn:    conn,
	}
}

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	q := PrivacyServiceSingleton.Queries

	exp.AddJSON("Assessments", "student/assessment.json", func() (any, error) {
		return q.GetAllAssessmentsByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Assessment Completions", "student/assessment_completion.json", func() (any, error) {
		return q.GetAllAssessmentCompletionsByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Evaluations", "student/evaluation.json", func() (any, error) {
		return q.GetAllEvaluationsByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Evaluation Completions", "student/evaluation_completion.json", func() (any, error) {
		return q.GetAllEvaluationCompletionsByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Action Items", "student/action_item.json", func() (any, error) {
		return q.GetAllActionItemsByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Feedback Items", "student/feedback_item.json", func() (any, error) {
		return q.GetAllFeedbackItemsByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})

	return nil
}
