package applicationDTO

import "github.com/google/uuid"

// ImportResult summarizes the outcome of a CSV student import. The import is all-or-nothing: any
// row error aborts the whole transaction and returns an error, so there is no per-row failure
// outcome here.
type ImportResult struct {
	Created int               `json:"created"`
	Updated int               `json:"updated"`
	Rows    []ImportRowResult `json:"rows"`
}

// ImportRowResult is the per-row outcome of a successful import.
type ImportRowResult struct {
	Index                 int        `json:"index"`
	UniversityLogin       string     `json:"universityLogin"`
	Outcome               string     `json:"outcome"` // "created" or "updated"
	CourseParticipationID *uuid.UUID `json:"courseParticipationId,omitempty"`
}
