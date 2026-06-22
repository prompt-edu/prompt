import { Assessment } from './assessment'
import { AssessmentCompletion } from './assessmentCompletion'
import { CategoryAssessment } from './categoryAssessment'
import { StudentScore } from './studentScore'
import { Evaluation } from './evaluation'

export interface StudentAssessment {
  courseParticipationID: string // UUID
  assessments: Assessment[]
  categoryAssessments: CategoryAssessment[]
  assessmentCompletion: AssessmentCompletion
  studentScore: StudentScore
  evaluations: Evaluation[]
}
