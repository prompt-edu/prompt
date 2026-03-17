package assessmentCompletionDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
)

type GradeWithParticipation struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Grade                 float64   `json:"grade"`
}

func GetGradesWithParticipationFromDBGradesWithParticipation(dbGrades []db.GetAllGradesRow) []GradeWithParticipation {
	gradesWithParticipation := make([]GradeWithParticipation, 0, len(dbGrades))
	for _, dbGrade := range dbGrades {
		gradesWithParticipation = append(gradesWithParticipation, GradeWithParticipation{
			CourseParticipationID: dbGrade.CourseParticipationID,
			Grade:                 utils.MapNumericToFloat64(dbGrade.GradeSuggestion),
		})
	}
	return gradesWithParticipation
}
