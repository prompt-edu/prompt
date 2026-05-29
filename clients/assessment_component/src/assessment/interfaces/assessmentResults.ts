import { Assessment } from './assessment'
import { AssessmentCompletion } from './assessmentCompletion'
import { ActionItem } from './actionItem'
import { CategoryAssessment } from './categoryAssessment'
import { StudentScore } from './studentScore'

export type AggregatedEvaluationResult = {
  competencyID: string
  averageScoreNumeric: number
}

export type StudentAssessmentResults = {
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
