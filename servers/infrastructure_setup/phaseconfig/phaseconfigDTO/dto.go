package phaseconfigDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// UpsertRequest is the request body for infrastructure setup phase configuration.
// Upstream team / participation data is now wired via the standard phase
// configurator, so the request only carries phase-local settings.
type UpsertRequest struct {
	SemesterTag string `json:"semesterTag"`
}

// Response is the API response for infrastructure setup phase configuration.
type Response struct {
	CoursePhaseID uuid.UUID `json:"coursePhaseId"`
	SemesterTag   string    `json:"semesterTag"`
}

// GetPhaseConfigDTOFromDBModel builds the API response from the DB row.
func GetPhaseConfigDTOFromDBModel(cfg db.CoursePhaseConfig) Response {
	return Response{
		CoursePhaseID: cfg.CoursePhaseID,
		SemesterTag:   cfg.SemesterTag,
	}
}
