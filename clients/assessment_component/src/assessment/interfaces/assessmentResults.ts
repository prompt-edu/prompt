import type { ActionItem } from './actionItem'
import type { Assessment } from './assessment'
import type { AssessmentCompletion } from './assessmentCompletion'
import type { CategoryAssessment } from './categoryAssessment'
import type { StudentScore } from './studentScore'

export interface AggregatedEvaluationResult {
  competencyID: string
  averageScoreNumeric: number
}

export interface StudentAssessmentResults {
  courseParticipationID: string
  coursePhaseID: string
  assessmentCompletion: AssessmentCompletion
  assessments: Assessment[]
  categoryAssessments: CategoryAssessment[]
  studentScore?: StudentScore
  peerEvaluationResults: AggregatedEvaluationResult[]
  selfEvaluationResults: AggregatedEvaluationResult[]
  actionItems?: ActionItem[]
}
