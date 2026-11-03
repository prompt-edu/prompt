package teamDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Tutor is one person to import as a tutor of a course phase. The university login identifies the
// tutor: an import sets exactly the listed teams for that login. CourseParticipationID is the id
// other phases evaluate the tutor under (the participation of a coach in the tutor course, or the
// Keycloak user id of a course editor). A tutor who already exists keeps the id they have.
type Tutor struct {
	CourseParticipationID uuid.UUID   `json:"courseParticipationID"`
	FirstName             string      `json:"firstName"`
	LastName              string      `json:"lastName"`
	UniversityLogin       string      `json:"universityLogin"`
	TeamIDs               []uuid.UUID `json:"teamIDs"`
}

func UniversityLoginParam(login string) pgtype.Text {
	if login == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: login, Valid: true}
}
