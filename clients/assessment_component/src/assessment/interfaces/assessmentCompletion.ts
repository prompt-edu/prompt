import type { CompetencyScoreCompletion } from './competencyScoreCompletion'

export type AssessmentCompletion = CompetencyScoreCompletion & {
  author: string
  comment: string
  gradeSuggestion: number
}

export type CreateOrUpdateAssessmentCompletionRequest = {
  courseParticipationID: string // UUID
  coursePhaseID: string // UUID
  author: string
  comment: string
  gradeSuggestion: number
  completed?: boolean
}
