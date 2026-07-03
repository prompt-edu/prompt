import type { ScoreLevel } from '@tumaet/prompt-shared-state'
import type { Assessment } from '../../../../interfaces/assessment'
import type { AssessmentCompletion } from '../../../../interfaces/assessmentCompletion'
import type { AssessmentParticipationWithStudent } from '../../../../interfaces/assessmentParticipationWithStudent'

export interface ParticipationWithAssessment {
  participation: AssessmentParticipationWithStudent
  assessments: Assessment[]
  scoreLevel: ScoreLevel
  scoreNumeric: number
  assessmentCompletion: AssessmentCompletion | undefined
}
