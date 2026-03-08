package teamDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
)

type Tutor struct {
	CoursePhaseID         uuid.UUID `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	FirstName             string    `json:"firstName"`
	LastName              string    `json:"lastName"`
	TeamID                uuid.UUID `json:"teamID"`
}

func GetTutorDTOFromDBModel(dbTutor db.Tutor) Tutor {
	return Tutor{
		CoursePhaseID:         dbTutor.CoursePhaseID,
		CourseParticipationID: dbTutor.CourseParticipationID,
		FirstName:             dbTutor.FirstName,
		LastName:              dbTutor.LastName,
		TeamID:                dbTutor.TeamID,
	}
}

func GetTutorDTOsFromDBModels(dbTutors []db.Tutor) []Tutor {
	tutors := make([]Tutor, 0, len(dbTutors))
	for _, dbTutor := range dbTutors {
		tutors = append(tutors, GetTutorDTOFromDBModel(dbTutor))
	}
	return tutors
}
