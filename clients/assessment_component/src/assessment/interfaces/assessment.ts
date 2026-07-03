import type { ScoreLevel } from '@tumaet/prompt-shared-state'
import type { CompetencyScore } from './competencyScore'

export interface Assessment extends CompetencyScore {
  assessedAt: string // ISO 8601 date string
  author: string
  authorID: string
}

export interface CreateOrUpdateAssessmentRequest {
  courseParticipationID: string // UUID
  coursePhaseID: string // UUID
  competencyID: string // UUID
  scoreLevel: ScoreLevel
}
