import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import type { ExtraParticipantColumn, ParticipantRow } from '@tumaet/prompt-ui-components'
import { compareNullableNumbers, matchesNumericRange } from '../utils/tableFns'

interface ScoreColumnOptions {
  id: string
  header: string
  getScore: (courseParticipationID: string) => number | null
}

export const createScoreColumn = (
  participations: CoursePhaseParticipationWithStudent[],
  { id, header, getScore }: ScoreColumnOptions,
): ExtraParticipantColumn<number | null> => ({
  id,
  header,

  accessorFn: (row: ParticipantRow) => getScore(row.courseParticipationID),

  cell: ({ getValue }) => getValue() ?? <span className='text-muted-foreground'>N/A</span>,

  enableSorting: true,
  sortFn: (rowA, rowB) =>
    compareNullableNumbers(
      getScore(rowA.original.courseParticipationID),
      getScore(rowB.original.courseParticipationID),
    ),

  enableColumnFilter: true,
  filterFn: (row, _columnId, filterValue) =>
    matchesNumericRange(getScore(row.original.courseParticipationID), filterValue),

  extraData: participations.map((p) => {
    const score = getScore(p.courseParticipationID)
    return {
      courseParticipationID: p.courseParticipationID,
      value: score,
      stringValue: score === null ? '' : String(score),
    }
  }),
})
