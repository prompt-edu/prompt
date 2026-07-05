import type { ScoreLevel } from '@tumaet/prompt-shared-state'
import type { AssessmentType } from './assessmentType'
import type { CompetencyScore } from './competencyScore'

export type Evaluation = CompetencyScore & {
  authorCourseParticipationID: string
  evaluatedAt: string
  type: AssessmentType
}

export type CreateOrUpdateEvaluationRequest = {
  courseParticipationID: string
  competencyID: string
  scoreLevel: ScoreLevel
  authorCourseParticipationID: string
  type: AssessmentType
}
