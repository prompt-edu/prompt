package surveyDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type PreferenceCount struct {
	Rank  int32 `json:"rank"`
	Count int32 `json:"count"`
}

type TeamPopularityStats struct {
	TeamID           uuid.UUID        `json:"teamId"`
	TeamName         string           `json:"teamName"`
	AvgPreference    float64          `json:"avgPreference"`
	ResponseCount    int64            `json:"responseCount"`
	PreferenceCounts []PreferenceCount `json:"preferenceCounts"`
}

type SkillDistributionStats struct {
	SkillID     uuid.UUID               `json:"skillId"`
	SkillName   string                  `json:"skillName"`
	LevelCounts map[db.SkillLevel]int64 `json:"levelCounts"`
}

type SurveyStatistics struct {
	TeamPopularityStatistics    []TeamPopularityStats    `json:"teamPopularityStatistics"`
	SkillDistributionStatistics []SkillDistributionStats `json:"skillDistributionStatistics"`
}

func GetSurveyStatisticsDTOFromDBModels(
	teamAvgRows []db.GetTeamPopularityStatisticsRow,
	teamCountRows []db.GetTeamPreferenceCountsRow,
	skillRows []db.GetSkillDistributionStatisticsRow,
) SurveyStatistics {
	prefCountMap := make(map[uuid.UUID][]PreferenceCount)
	for _, row := range teamCountRows {
		prefCountMap[row.TeamID] = append(prefCountMap[row.TeamID], PreferenceCount{
			Rank:  row.Preference,
			Count: row.Count,
		})
	}

	teamStats := make([]TeamPopularityStats, 0, len(teamAvgRows))
	for _, row := range teamAvgRows {
		teamStats = append(teamStats, TeamPopularityStats{
			TeamID:           row.TeamID,
			TeamName:         row.TeamName,
			AvgPreference:    row.AvgPreference,
			ResponseCount:    row.ResponseCount,
			PreferenceCounts: prefCountMap[row.TeamID],
		})
	}

	skillMap := make(map[uuid.UUID]*SkillDistributionStats)
	for _, row := range skillRows {
		if _, exists := skillMap[row.SkillID]; !exists {
			skillMap[row.SkillID] = &SkillDistributionStats{
				SkillID:     row.SkillID,
				SkillName:   row.SkillName,
				LevelCounts: make(map[db.SkillLevel]int64),
			}
		}
		skillMap[row.SkillID].LevelCounts[row.SkillLevel] = row.Count
	}

	skillStats := make([]SkillDistributionStats, 0, len(skillMap))
	for _, v := range skillMap {
		skillStats = append(skillStats, *v)
	}

	return SurveyStatistics{
		TeamPopularityStatistics:    teamStats,
		SkillDistributionStatistics: skillStats,
	}
}
