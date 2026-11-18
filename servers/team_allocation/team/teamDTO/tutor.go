package teamDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
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

// TutorAssignment is one team a tutor leads. ID is the id the tutor is stored and evaluated under,
// and UniversityLogin (empty for a row stored without one) identifies the tutor across import
// sources, so the import dialog can start a person from the teams they already lead.
type TutorAssignment struct {
	ID              uuid.UUID `json:"id"`
	UniversityLogin string    `json:"universityLogin"`
	TeamID          uuid.UUID `json:"teamID"`
}

func GetTutorAssignmentsFromDBModels(rows []db.GetTutorAssignmentsByCoursePhaseRow) []TutorAssignment {
	assignments := make([]TutorAssignment, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, TutorAssignment{
			ID:              row.CourseParticipationID,
			UniversityLogin: row.UniversityLogin.String,
			TeamID:          row.TeamID,
		})
	}
	return assignments
}
