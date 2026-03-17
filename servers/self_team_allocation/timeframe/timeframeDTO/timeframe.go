package timeframeDTO

import (
	"time"

	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
)

type Timeframe struct {
	TimeframeSet bool      `json:"timeframeSet"`
	StartTime    time.Time `json:"startTime" binding:"required"`
	EndTime      time.Time `json:"endTime" binding:"required"`
}

func GetTimeframeDTOFromDBModel(timeframe db.GetTimeframeRow) Timeframe {
	return Timeframe{
		TimeframeSet: true,
		StartTime:    timeframe.Starttime.Time,
		EndTime:      timeframe.Endtime.Time,
	}
}
