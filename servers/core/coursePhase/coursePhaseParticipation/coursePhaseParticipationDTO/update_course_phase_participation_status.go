package coursePhaseParticipationDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type UpdateCoursePhaseParticipationStatus struct {
	PassStatus             db.PassStatus `json:"passStatus"`
	CourseParticipationIDs []uuid.UUID   `json:"courseParticipationIDs"`
}
