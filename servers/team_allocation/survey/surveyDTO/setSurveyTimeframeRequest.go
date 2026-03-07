package surveyDTO

import (
	"time"

	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type SurveyTimeframe struct {
	TimeframeSet   bool      `json:"timeframeSet"`
	SurveyStart    time.Time `json:"surveyStart" binding:"required"`
	SurveyDeadline time.Time `json:"surveyDeadline" binding:"required"`
}

func GetSurveyTimeframeDTOFromDBModel(timeframe db.GetSurveyTimeframeRow) SurveyTimeframe {
	return SurveyTimeframe{
		TimeframeSet:   true,
		SurveyStart:    timeframe.SurveyStart.Time,
		SurveyDeadline: timeframe.SurveyDeadline.Time,
	}
}
