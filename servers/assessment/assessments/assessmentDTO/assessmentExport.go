package assessmentDTO

import (
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
)

type AssessmentExport struct {
	ExportedAt            time.Time                  `json:"exportedAt"`
	CoursePhaseID         uuid.UUID                  `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID                  `json:"courseParticipationID"`
	StudentAssessment     StudentAssessment          `json:"studentAssessment"`
	ActionItems           []actionItemDTO.ActionItem `json:"actionItems"`
}
