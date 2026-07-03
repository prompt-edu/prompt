import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'

export type AssessmentParticipationWithStudent = CoursePhaseParticipationWithStudent & {
  teamID: string
}
