package studentDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type Student struct {
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

func GetStudentDTOFromDBModel(model db.Student) Student {
	return Student{
		ID:                   model.ID,
		FirstName:            model.FirstName.String,
		LastName:             model.LastName.String,
		Email:                model.Email.String,
		MatriculationNumber:  model.MatriculationNumber.String,
		UniversityLogin:      model.UniversityLogin.String,
		HasUniversityAccount: model.HasUniversityAccount.Bool,
		Gender:               model.Gender,
		Nationality:          model.Nationality.String,
		StudyDegree:          model.StudyDegree,
		StudyProgram:         model.StudyProgram.String,
		CurrentSemester:      model.CurrentSemester,
	}
}

func GetStudentDTOFromApplicationParticipation(model db.GetAllApplicationParticipationsRow) Student {
	return Student{
		ID:                   model.StudentID,
		FirstName:            model.FirstName.String,
		LastName:             model.LastName.String,
		Email:                model.Email.String,
		MatriculationNumber:  model.MatriculationNumber.String,
		UniversityLogin:      model.UniversityLogin.String,
		HasUniversityAccount: model.HasUniversityAccount.Bool,
		Gender:               model.Gender,
		Nationality:          model.Nationality.String,
		StudyDegree:          model.StudyDegree,
		StudyProgram:         model.StudyProgram.String,
		CurrentSemester:      model.CurrentSemester,
	}
}

func GetStudentDTOFromCourseParticipation(model db.GetStudentsOfCoursePhaseRow) Student {
	return Student{
		ID:                   model.ID,
		FirstName:            model.FirstName.String,
		LastName:             model.LastName.String,
		Email:                model.Email.String,
		MatriculationNumber:  model.MatriculationNumber.String,
		UniversityLogin:      model.UniversityLogin.String,
		HasUniversityAccount: model.HasUniversityAccount.Bool,
		Gender:               model.Gender,
		Nationality:          model.Nationality.String,
		StudyDegree:          model.StudyDegree,
		StudyProgram:         model.StudyProgram.String,
		CurrentSemester:      model.CurrentSemester,
	}

}

func (s Student) GetDBModel() db.Student {
	return db.Student{
		ID:                   s.ID,
		FirstName:            pgtype.Text{String: s.FirstName, Valid: true},
		LastName:             pgtype.Text{String: s.LastName, Valid: true},
		Email:                pgtype.Text{String: s.Email, Valid: true},
		MatriculationNumber:  pgtype.Text{String: s.MatriculationNumber, Valid: true},
		UniversityLogin:      pgtype.Text{String: s.UniversityLogin, Valid: true},
		HasUniversityAccount: pgtype.Bool{Bool: s.HasUniversityAccount, Valid: true},
		Gender:               s.Gender,
		Nationality:          pgtype.Text{String: s.Nationality, Valid: true},
		StudyDegree:          s.StudyDegree,
		StudyProgram:         pgtype.Text{String: s.StudyProgram, Valid: true},
		CurrentSemester:      s.CurrentSemester,
	}
}
