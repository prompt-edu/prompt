export type SkillLevel = 'very_bad' | 'bad' | 'ok' | 'good' | 'very_good'

export interface PreferenceCount {
  rank: number
  count: number
}

export interface TeamPopularityStats {
  teamId: string
  teamName: string
  avgPreference: number | null // null when no responses yet
  responseCount: number
  preferenceCounts: PreferenceCount[]
}

export interface SkillDistributionStats {
  skillId: string
  skillName: string
  levelCounts: Partial<Record<SkillLevel, number>>
}

export interface SurveyStatistics {
  teamPopularityStatistics: TeamPopularityStats[]
  skillDistributionStatistics: SkillDistributionStats[]
}
