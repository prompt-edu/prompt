import { Team } from '@tumaet/prompt-shared-state'

import { AssessmentParticipationWithStudent } from '../../../interfaces/assessmentParticipationWithStudent'
import { EvaluationCompletion } from '../../../interfaces/evaluationCompletion'

import { PeerEvaluationCompletionBadge } from '../../components/badges'

import { createEvaluationLookup, getEvaluationCounts } from '../utils/evaluationUtils'
import { ExtraParticipantColumn } from '@/components/pages/CoursePhaseParticipationsTable/table/participationRow'

export const createTutorEvalStatusColumn = (
  tutorEvaluationCompletions: EvaluationCompletion[] | undefined,
  teams: Team[],
  participations: AssessmentParticipationWithStudent[],
  isEnabled: boolean,
): ExtraParticipantColumn<{ completed: number; total: number }> | undefined => {
  if (!isEnabled) return undefined

  const evaluationLookup = createEvaluationLookup(tutorEvaluationCompletions)

  const getCountsForParticipation = (courseParticipationID: string) => {
    const team = teams.find((t) => t.members.some((member) => member.id === courseParticipationID))

    if (!team) {
      return { completed: 0, total: 0 }
    }

    const tutorIds = team.tutors
      .map((tutor) => tutor.id)
      .filter((id): id is string => id !== undefined)

    return getEvaluationCounts(courseParticipationID, tutorIds, evaluationLookup)
  }

  return {
    id: 'tutorEvalStatus',
    header: 'Tutor Eval',

    accessorFn: (row) => getCountsForParticipation(row.courseParticipationID),

    cell: ({ getValue }) => {
      const { completed, total } = getValue()
      return <PeerEvaluationCompletionBadge completed={completed} total={total} />
    },

    enableSorting: true,
    sortingFn: (rowA, rowB) => {
      const a = getCountsForParticipation(rowA.original.courseParticipationID)
      const b = getCountsForParticipation(rowB.original.courseParticipationID)

      const ratioA = a.total > 0 ? a.completed / a.total : 0
      const ratioB = b.total > 0 ? b.completed / b.total : 0

      return ratioA - ratioB
    },

    extraData: participations.map((p) => {
      const counts = getCountsForParticipation(p.courseParticipationID)

      return {
        courseParticipationID: p.courseParticipationID,
        value: counts,
        stringValue: `${counts.completed}/${counts.total}`,
      }
    }),
  }
}
