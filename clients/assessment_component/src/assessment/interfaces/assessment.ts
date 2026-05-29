import { CompetencyScore } from './competencyScore'
import { ScoreLevel } from '@tumaet/prompt-shared-state'

export type Assessment = CompetencyScore & {
  assessedAt: string // ISO 8601 date string
  author: string
  authorID: string
}

export type CreateOrUpdateAssessmentRequest = {
  courseParticipationID: string // UUID
  coursePhaseID: string // UUID
  competencyID: string // UUID
  scoreLevel: ScoreLevel
  author: string
  authorID: string
}
