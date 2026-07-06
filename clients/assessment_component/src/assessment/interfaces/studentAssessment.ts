import type { Assessment } from './assessment'
import type { AssessmentCompletion } from './assessmentCompletion'
import type { CategoryAssessment } from './categoryAssessment'
import type { Evaluation } from './evaluation'
import type { StudentScore } from './studentScore'

export interface StudentAssessment {
  courseParticipationID: string // UUID
  assessments: Assessment[]
  categoryAssessments: CategoryAssessment[]
  assessmentCompletion: AssessmentCompletion
  studentScore: StudentScore
  evaluations: Evaluation[]
}
