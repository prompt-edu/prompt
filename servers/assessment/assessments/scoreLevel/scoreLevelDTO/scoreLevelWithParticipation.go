package scoreLevelDTO

import (
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
)

type ScoreLevelWithParticipation struct {
	CourseParticipationID string     `json:"courseParticipationID"`
	ScoreLevel            ScoreLevel `json:"scoreLevel"`
	ScoreNumeric          float64    `json:"scoreNumeric"`
}

func GetScoreLevelsFromDBScoreLevels(scoreLevels []db.GetAllScoreLevelsRow) []ScoreLevelWithParticipation {
	scoreLevelWithParticipation := make([]ScoreLevelWithParticipation, 0, len(scoreLevels))
	for _, scoreLevel := range scoreLevels {
		scoreLevelWithParticipation = append(scoreLevelWithParticipation, ScoreLevelWithParticipation{
			CourseParticipationID: scoreLevel.CourseParticipationID.String(),
			ScoreLevel:            MapDBScoreLevelToDTO(db.ScoreLevel(scoreLevel.ScoreLevel)),
			ScoreNumeric:          utils.MapNumericToFloat64(scoreLevel.ScoreNumeric),
		})
	}
	return scoreLevelWithParticipation
}

func MapToScoreLevelWithParticipation(scoreLevel db.GetAllScoreLevelsRow) ScoreLevelWithParticipation {
	return ScoreLevelWithParticipation{
		CourseParticipationID: scoreLevel.CourseParticipationID.String(),
		ScoreLevel:            MapDBScoreLevelToDTO(db.ScoreLevel(scoreLevel.ScoreLevel)),
	}
}
