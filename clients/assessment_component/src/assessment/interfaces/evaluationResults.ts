import type { ScoreLevel } from '@tumaet/prompt-shared-state'

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
