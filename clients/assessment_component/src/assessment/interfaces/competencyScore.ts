import type { ScoreLevel } from '@tumaet/prompt-shared-state'

export type CompetencyScore = {
  id: string // UUID
  courseParticipationID: string // UUID
  coursePhaseID: string // UUID
  competencyID: string // UUID
  scoreLevel: ScoreLevel
}
