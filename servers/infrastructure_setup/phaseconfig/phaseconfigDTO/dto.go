package phaseconfigDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// UpsertRequest is the request body for infrastructure setup phase configuration.
type UpsertRequest struct {
	TeamSourceCoursePhaseID    *uuid.UUID `json:"teamSourceCoursePhaseId"`
	StudentSourceCoursePhaseID *uuid.UUID `json:"studentSourceCoursePhaseId"`
	SemesterTag                string     `json:"semesterTag"`
}

// Response is the API response for infrastructure setup phase configuration.
type Response struct {
	CoursePhaseID              uuid.UUID  `json:"coursePhaseId"`
	TeamSourceCoursePhaseID    *uuid.UUID `json:"teamSourceCoursePhaseId,omitempty"`
	StudentSourceCoursePhaseID *uuid.UUID `json:"studentSourceCoursePhaseId,omitempty"`
	SemesterTag                string     `json:"semesterTag"`
}

// GetPhaseConfigDTOFromDBModel builds the API response from the DB row.
func GetPhaseConfigDTOFromDBModel(cfg db.CoursePhaseConfig) Response {
	return Response{
		CoursePhaseID:              cfg.CoursePhaseID,
		TeamSourceCoursePhaseID:    cfg.TeamSourceCoursePhaseID,
		StudentSourceCoursePhaseID: cfg.StudentSourceCoursePhaseID,
		SemesterTag:                cfg.SemesterTag,
	}
}
