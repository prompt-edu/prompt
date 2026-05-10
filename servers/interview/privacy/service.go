package privacy

import (
	"github.com/gin-gonic/gin"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

type InterviewPrivacyService struct {
	queries db.Queries
}

var singleton *InterviewPrivacyService

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Interview Assignments", "interview_assignments.json", func() (any, error) {
		return singleton.queries.GetInterviewAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs)
	})
	return nil
}
