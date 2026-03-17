import { AssessmentCompletion } from '../../../interfaces/assessmentCompletion'
import { GradeSuggestionBadge } from '../../components/badges'
import {
  ExtraParticipantColumn,
  ParticipantRow,
} from '@/components/pages/CoursePhaseParticipationsTable/table/participationRow'

export const createGradeSuggestionColumn = (
  assessmentCompletions: AssessmentCompletion[] | undefined,
): ExtraParticipantColumn<number | null> | undefined => {
  if (!assessmentCompletions) return undefined

  const completedGradings = assessmentCompletions.filter((a) => a.completed)

  const getGradeForParticipation = (courseParticipationID: string): number | null => {
    const match = completedGradings.find((a) => a.courseParticipationID === courseParticipationID)
    return match ? match.gradeSuggestion : null
  }

  return {
    id: 'gradeSuggestion',
    header: 'Grade',

    accessorFn: (row: ParticipantRow) => getGradeForParticipation(row.courseParticipationID),

    cell: ({ getValue }) => {
      const grade = getValue()
      return grade !== null ? <GradeSuggestionBadge gradeSuggestion={grade} text={false} /> : null
    },

    enableSorting: true,
    sortingFn: (rowA, rowB) => {
      const a = getGradeForParticipation(rowA.original.courseParticipationID) ?? 6
      const b = getGradeForParticipation(rowB.original.courseParticipationID) ?? 6
      return a - b
    },

    extraData: completedGradings.map((s) => ({
      courseParticipationID: s.courseParticipationID,
      value: s.gradeSuggestion,
      stringValue: s.gradeSuggestion.toFixed(1),
    })),
  }
}
