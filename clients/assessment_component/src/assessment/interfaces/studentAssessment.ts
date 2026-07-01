import { Assessment } from './assessment'
import { AssessmentCompletion } from './assessmentCompletion'
import { CategoryAssessment } from './categoryAssessment'
import { Evaluation } from './evaluation'
import { StudentScore } from './studentScore'

export interface StudentAssessment {
  courseParticipationID: string // UUID
  assessments: Assessment[]
  categoryAssessments: CategoryAssessment[]
  assessmentCompletion: AssessmentCompletion
  studentScore: StudentScore
  evaluations: Evaluation[]
}
