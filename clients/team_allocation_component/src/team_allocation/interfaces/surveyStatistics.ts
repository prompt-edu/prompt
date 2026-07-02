import { SkillLevel } from './skillResponse'

export interface PreferenceCount {
  rank: number
  count: number
}

export interface TeamPopularityStats {
  teamId: string
  teamName: string
  avgPreference: number | null
  responseCount: number
  preferenceCounts: PreferenceCount[]
}

export interface SkillDistributionStats {
  skillId: string
  skillName: string
  levelCounts: Partial<Record<SkillLevel, number>>
}

export interface SurveyStatistics {
  respondentCount: number
  teamPopularityStatistics: TeamPopularityStats[]
  skillDistributionStatistics: SkillDistributionStats[]
}
