// Package privacy implements the SDK privacy export and deletion endpoints for the
// infrastructure setup phase.
package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// PrivacyService answers core's privacy export and deletion requests.
//
// What the service stores about a person is the resource instances of its per_student
// configs: which external resource was provisioned for them, its ID and URL, and why a
// run failed. Provider and resource configs are phase configuration, and a team-scoped
// instance belongs to the team, so neither is subject data.
type PrivacyService struct {
	queries *db.Queries
}

// NewPrivacyService creates a PrivacyService.
func NewPrivacyService(pool *pgxpool.Pool) *PrivacyService {
	return &PrivacyService{queries: db.New(pool)}
}

// DataExportHandler implements promptTypes.PrivacyDataExportHandler.
func (s *PrivacyService) DataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Provisioned Resources", "infrastructure_setup.json", func() (any, error) {
		return s.queries.GetResourceInstancesByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	return nil
}

// DataDeletionHandler implements promptTypes.PrivacyDataDeletionHandler.
//
// Only PROMPT's own records go. The external resources are never deleted by this phase
// (a provider adopts by name, so it cannot tell what belongs to the course), which is
// also true here: a GitLab group named after the student survives the deletion and has
// to be removed in the provider by hand.
func (s *PrivacyService) DataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	return s.queries.DeleteResourceInstancesByCourseParticipationIDs(c, subject.CourseParticipationIDs)
}
