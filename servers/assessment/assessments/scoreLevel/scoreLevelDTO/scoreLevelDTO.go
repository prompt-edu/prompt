package scoreLevelDTO

import (
	"log"

	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type ScoreLevel string

const (
	ScoreLevelVeryBad  ScoreLevel = "veryBad"
	ScoreLevelBad      ScoreLevel = "bad"
	ScoreLevelOk       ScoreLevel = "ok"
	ScoreLevelGood     ScoreLevel = "good"
	ScoreLevelVeryGood ScoreLevel = "veryGood"
)

func MapDBScoreLevelToDTO(scoreLevel db.ScoreLevel) ScoreLevel {
	switch scoreLevel {
	case db.ScoreLevelVeryBad:
		return ScoreLevelVeryBad
	case db.ScoreLevelBad:
		return ScoreLevelBad
	case db.ScoreLevelOk:
		return ScoreLevelOk
	case db.ScoreLevelGood:
		return ScoreLevelGood
	case db.ScoreLevelVeryGood:
		return ScoreLevelVeryGood
	}

	log.Println("Warning: Unrecognized score level in MapDBScoreLevelToDTO:", scoreLevel)
	return ScoreLevelVeryBad // Default case, should not happen
}

func MapDTOtoDBScoreLevel(scoreLevel ScoreLevel) db.ScoreLevel {
	switch scoreLevel {
	case ScoreLevelVeryBad:
		return db.ScoreLevelVeryBad
	case ScoreLevelBad:
		return db.ScoreLevelBad
	case ScoreLevelOk:
		return db.ScoreLevelOk
	case ScoreLevelGood:
		return db.ScoreLevelGood
	case ScoreLevelVeryGood:
		return db.ScoreLevelVeryGood
	}

	log.Println("Warning: Unrecognized score level in MapDTOtoDBScoreLevel:", scoreLevel)
	return db.ScoreLevelVeryBad // Default case, should not happen
}

func MapScoreLevelToNumber(level ScoreLevel) float64 {
	switch level {
	case ScoreLevelVeryGood:
		return 1
	case ScoreLevelGood:
		return 2
	case ScoreLevelOk:
		return 3
	case ScoreLevelBad:
		return 4
	default:
		return 5
	}
}
