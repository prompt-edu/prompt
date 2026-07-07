package teamDTO

import (
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type Tutor struct {
	CoursePhaseID         uuid.UUID `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	FirstName             string    `json:"firstName"`
	LastName              string    `json:"lastName"`
	TeamID                uuid.UUID `json:"teamID"`
	UniversityLogin       string    `json:"universityLogin"`
}

type UpdateTutorTeamRequest struct {
	TeamID uuid.UUID `json:"teamID" binding:"required"`
}

func GetTutorDTOFromDBModel(dbTutor db.Tutor) Tutor {
	return Tutor{
		CoursePhaseID:         dbTutor.CoursePhaseID,
		CourseParticipationID: dbTutor.CourseParticipationID,
		FirstName:             dbTutor.FirstName,
		LastName:              dbTutor.LastName,
		TeamID:                dbTutor.TeamID,
		UniversityLogin:       dbTutor.UniversityLogin.String,
	}
}

func NormalizeUniversityLogin(login string) string {
	return strings.TrimSpace(strings.ToLower(login))
}

func UniversityLoginParam(login string) pgtype.Text {
	if login == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: login, Valid: true}
}
