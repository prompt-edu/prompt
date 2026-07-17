package applicationDTO

import "github.com/google/uuid"

// ImportResult summarizes the outcome of a CSV student import.
type ImportResult struct {
	Created int               `json:"created"`
	Updated int               `json:"updated"`
	Failed  int               `json:"failed"`
	Rows    []ImportRowResult `json:"rows"`
}

// ImportRowResult is the per-row outcome of an import.
type ImportRowResult struct {
	Index                 int        `json:"index"`
	UniversityLogin       string     `json:"universityLogin"`
	Outcome               string     `json:"outcome"` // "created", "updated" or "failed"
	Reason                string     `json:"reason,omitempty"`
	CourseParticipationID *uuid.UUID `json:"courseParticipationId,omitempty"`
}
