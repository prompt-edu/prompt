package studentDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type CreateStudent struct {
	ID                   uuid.UUID      `json:"id"`
	FirstName            string         `json:"firstName"`
	LastName             string         `json:"lastName"`
	Email                string         `json:"email"`
	MatriculationNumber  string         `json:"matriculationNumber"`
	UniversityLogin      string         `json:"universityLogin"`
	HasUniversityAccount bool           `json:"hasUniversityAccount"`
	Gender               db.Gender      `json:"gender"`
	Nationality          string         `json:"nationality"`
	StudyDegree          db.StudyDegree `json:"studyDegree"`
	StudyProgram         string         `json:"studyProgram"`
	CurrentSemester      pgtype.Int4    `json:"currentSemester" swaggertype:"integer"`
}

func (c CreateStudent) GetDBModel() db.CreateStudentParams {
	return db.CreateStudentParams{
		FirstName:            pgtype.Text{String: c.FirstName, Valid: true},
		LastName:             pgtype.Text{String: c.LastName, Valid: true},
		Email:                pgtype.Text{String: c.Email, Valid: true},
		MatriculationNumber:  pgtype.Text{String: c.MatriculationNumber, Valid: true},
		UniversityLogin:      pgtype.Text{String: c.UniversityLogin, Valid: true},
		HasUniversityAccount: pgtype.Bool{Bool: c.HasUniversityAccount, Valid: true},
		Gender:               c.Gender,
		Nationality:          pgtype.Text{String: c.Nationality, Valid: true},
		StudyDegree:          c.StudyDegree,
		StudyProgram:         pgtype.Text{String: c.StudyProgram, Valid: true},
		CurrentSemester:      c.CurrentSemester,
	}
}
