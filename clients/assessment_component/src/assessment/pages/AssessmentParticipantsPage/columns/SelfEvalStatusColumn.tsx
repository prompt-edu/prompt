import { EvaluationCompletion } from '../../../interfaces/evaluationCompletion'
import { AssessmentStatusBadge } from '../../components/badges'
import {
  ExtraParticipantColumn,
  ParticipantRow,
} from '@/components/pages/CoursePhaseParticipationsTable/table/participationRow'

export const createSelfEvalStatusColumn = (
  selfEvaluationCompletions: EvaluationCompletion[] | undefined,
  isEnabled: boolean,
): ExtraParticipantColumn<boolean> | undefined => {
  if (!isEnabled) return undefined

  const isCompleted = (courseParticipationID: string): boolean => {
    return (
      selfEvaluationCompletions?.some(
        (s) => s.courseParticipationID === courseParticipationID && s.completed,
      ) ?? false
    )
  }

  return {
    id: 'selfEvalStatus',
    header: 'Self Eval',

    accessorFn: (row: ParticipantRow) => isCompleted(row.courseParticipationID),

    cell: ({ getValue }) =>
      getValue() ? <AssessmentStatusBadge remainingAssessments={0} isFinalized={true} /> : null,

    enableSorting: true,
    sortingFn: (rowA, rowB) =>
      Number(isCompleted(rowA.original.courseParticipationID)) -
      Number(isCompleted(rowB.original.courseParticipationID)),

    extraData:
      selfEvaluationCompletions?.map((s) => ({
        courseParticipationID: s.courseParticipationID,
        value: s.completed,
        stringValue: s.completed ? 'Yes' : 'No',
      })) ?? [],
  }
}
