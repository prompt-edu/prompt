import type { ScoreLevel } from '@tumaet/prompt-shared-state'
import type { FeedbackType } from './feedbackItem'

export interface SelfEvaluationResult {
  competencyID: string
  scoreLevel: ScoreLevel
}

export interface AggregatedEvaluationResult {
  competencyID: string
  averageScoreNumeric: number
}

export interface StudentEvaluationResults {
  courseParticipationID: string
  coursePhaseID: string
  selfResults: SelfEvaluationResult[]
  peerResults: AggregatedEvaluationResult[]
}

export interface AnonymousFeedbackItem {
  feedbackType: FeedbackType
  feedbackText: string
}

export interface TutorEvaluationResults {
  tutorID: string
  coursePhaseID: string
  evaluatorCount: number
  results: AggregatedEvaluationResult[]
  feedbackItems: AnonymousFeedbackItem[]
}
