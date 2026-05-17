package phaseconfig

import "github.com/google/uuid"

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
