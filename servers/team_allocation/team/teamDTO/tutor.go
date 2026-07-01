package teamDTO

import (
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
	UniversityLogin       string    `json:"universityLogin" binding:"omitempty"`
}

type UpdateTutorTeamRequest struct {
	TeamID uuid.UUID `json:"teamID" binding:"required"`
}

func GetTutorDTOFromDBModel(dbTutor db.Tutor) Tutor {
	universityLogin := ""
	if dbTutor.UniversityLogin.Valid {
		universityLogin = dbTutor.UniversityLogin.String
	}
	return Tutor{
		CoursePhaseID:         dbTutor.CoursePhaseID,
		CourseParticipationID: dbTutor.CourseParticipationID,
		FirstName:             dbTutor.FirstName,
		LastName:              dbTutor.LastName,
		TeamID:                dbTutor.TeamID,
		UniversityLogin:       universityLogin,
	}
}

func UniversityLoginParam(login string) pgtype.Text {
	if login == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: login, Valid: true}
}
