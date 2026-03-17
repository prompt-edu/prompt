import { Team } from '@tumaet/prompt-shared-state'
import { AssessmentParticipationWithStudent } from '../../../interfaces/assessmentParticipationWithStudent'
import {
  ExtraParticipantColumn,
  ParticipantRow,
} from '@/components/pages/CoursePhaseParticipationsTable/table/participationRow'

export const createTeamColumn = (
  teams: Team[],
  participations: AssessmentParticipationWithStudent[],
): ExtraParticipantColumn<string> | undefined => {
  if (teams.length === 0) return undefined

  const getTeamName = (courseParticipationID: string): string => {
    return (
      teams.find((t) => t.members.some((member) => member.id === courseParticipationID))?.name ?? ''
    )
  }

  return {
    id: 'team',
    header: 'Team',

    accessorFn: (row: ParticipantRow) => getTeamName(row.courseParticipationID),

    cell: ({ getValue }) => {
      const teamName = getValue<string>()
      return teamName || null
    },

    enableSorting: true,

    enableColumnFilter: true,

    sortingFn: (rowA, rowB) =>
      getTeamName(rowA.original.courseParticipationID).localeCompare(
        getTeamName(rowB.original.courseParticipationID),
      ),

    extraData: participations.map((p) => {
      const teamName = getTeamName(p.courseParticipationID)
      return {
        courseParticipationID: p.courseParticipationID,
        value: teamName,
        stringValue: teamName,
      }
    }),

    filterFn: (row, _columnId, filterValue) => {
      if (!Array.isArray(filterValue)) return false
      return filterValue.includes(getTeamName(row.original.courseParticipationID))
    },
  }
}
