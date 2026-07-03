import type { ScoreLevel } from '@tumaet/prompt-shared-state'

export interface ScoreDistributionDataPoint {
  shortLabel: string
  label: string
  average: number
  lowerQuartile: number
  median: ScoreLevel
  upperQuartile: number
  counts: Record<ScoreLevel, number>
}
