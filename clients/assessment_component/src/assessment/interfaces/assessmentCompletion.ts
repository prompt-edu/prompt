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

export type AssessmentCompletionSkipReason =
  | 'no_completion'
  | 'remaining_assessments'
  | 'already_completed'
  | 'not_completed'

export interface SkippedAssessmentCompletion {
  courseParticipationID: string // UUID
  reason: AssessmentCompletionSkipReason
}

export interface BatchMarkAssessmentCompletionsResult {
  marked: string[] // UUIDs
  skipped: SkippedAssessmentCompletion[]
}

export interface BatchUnmarkAssessmentCompletionsResult {
  unmarked: string[] // UUIDs
  skipped: SkippedAssessmentCompletion[]
}
