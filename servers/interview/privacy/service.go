package privacy

import (
	"github.com/gin-gonic/gin"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

type PrivacyService struct {
	Queries db.Queries
}

var PrivacyServiceSingleton *PrivacyService

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Interview Assignments", "interview_assignments.json", func() (any, error) {
		return PrivacyServiceSingleton.Queries.GetInterviewAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs)
	})
	return nil
}

func PrivacyDataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	return PrivacyServiceSingleton.Queries.DeleteInterviewAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs)
}
