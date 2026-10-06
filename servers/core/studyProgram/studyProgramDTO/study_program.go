package studyProgramDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type StudyProgram struct {
	ID        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	ShortName pgtype.Text `json:"shortName" swaggertype:"string"`
}

type CreateStudyProgram struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

type UpdateStudyProgram struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

type StudyProgramStudentCount struct {
	StudyProgramID uuid.UUID `json:"studyProgramID"`
	StudentCount   int64     `json:"studentCount"`
}

func StudyProgramFromDBModel(model db.StudyProgram) StudyProgram {
	return StudyProgram{
		ID:        model.ID,
		Name:      model.Name,
		ShortName: model.ShortName,
	}
}

func StudyProgramsFromDBModels(models []db.StudyProgram) []StudyProgram {
	studyPrograms := make([]StudyProgram, 0, len(models))
	for _, model := range models {
		studyPrograms = append(studyPrograms, StudyProgramFromDBModel(model))
	}
	return studyPrograms
}

func StudentCountsFromDBModels(rows []db.CountStudentsPerStudyProgramRow) []StudyProgramStudentCount {
	counts := make([]StudyProgramStudentCount, 0, len(rows))
	for _, row := range rows {
		counts = append(counts, StudyProgramStudentCount{
			StudyProgramID: row.StudyProgramID,
			StudentCount:   row.StudentCount,
		})
	}
	return counts
}
