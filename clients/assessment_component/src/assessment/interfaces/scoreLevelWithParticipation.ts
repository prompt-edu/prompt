import type { ScoreLevel } from '@tumaet/prompt-shared-state'

export type ScoreLevelWithParticipation = {
  courseParticipationID: string
  scoreLevel: ScoreLevel
  scoreNumeric: number
}
