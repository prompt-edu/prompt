import { ActionItem } from './actionItem'
import { Assessment } from './assessment'
import { AssessmentCompletion } from './assessmentCompletion'
import { CategoryAssessment } from './categoryAssessment'
import { StudentScore } from './studentScore'

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
  studentScore: StudentScore
  peerEvaluationResults: AggregatedEvaluationResult[]
  selfEvaluationResults: AggregatedEvaluationResult[]
  actionItems?: ActionItem[]
}
