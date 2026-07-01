import { ScoreLevel } from '@tumaet/prompt-shared-state'
import { Assessment } from '../../../../interfaces/assessment'
import { AssessmentCompletion } from '../../../../interfaces/assessmentCompletion'
import { AssessmentParticipationWithStudent } from '../../../../interfaces/assessmentParticipationWithStudent'

export interface ParticipationWithAssessment {
  participation: AssessmentParticipationWithStudent
  assessments: Assessment[]
  scoreLevel: ScoreLevel
  scoreNumeric: number
  assessmentCompletion: AssessmentCompletion | undefined
}
