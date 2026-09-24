import { useQuery } from '@tanstack/react-query'
import { type Team, useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  type ExtraParticipantColumn,
  ManagementPageHeader,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { getAllTeams } from '../../network/queries/getAllTeams'

export const SelfTeamAllocationParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const participationsQuery = useGetCoursePhaseParticipants()

  const teamsQuery = useQuery<Team[]>({
    queryKey: ['self_team_allocations', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
  })

  const coursePhaseParticipations = participationsQuery.data
  const teams = teamsQuery.data

  const extraColumns: ExtraParticipantColumn<string>[] = useMemo(() => {
    if (!teams) return []

    const teamNameByParticipation = new Map<string, string>()

    for (const team of teams) {
      for (const member of team.members) {
        if (member.id) {
          teamNameByParticipation.set(member.id, team.name)
        }
      }
    }

    const teamNameExtraData =
      coursePhaseParticipations?.participations?.map((participation) => {
        const teamName =
          teamNameByParticipation.get(participation.courseParticipationID) ?? 'No Team'

        return {
          courseParticipationID: participation.courseParticipationID,
          value: teamName,
          stringValue: teamName,
        }
      }) ?? []

    return [
      {
        id: 'allocatedTeam',
        header: 'Allocated Team',

        accessorFn: (row) => teamNameByParticipation.get(row.courseParticipationID) ?? 'No Team',

        cell: ({ getValue }) => getValue(),

        extraData: teamNameExtraData,
        enableSorting: true,
        sortingFn: (rowA, rowB) => {
          const a = rowA.getValue('allocatedTeam') as string
          const b = rowB.getValue('allocatedTeam') as string
          return a.localeCompare(b)
        },
        enableColumnFilter: true,
        filterFn: (row, columnId, filterValue) => {
          const value = String(row.getValue(columnId) ?? '').toLowerCase()
          if (!Array.isArray(filterValue)) return false
          return filterValue.map((v) => v.toLowerCase()).includes(value)
        },
      },
    ]
  }, [coursePhaseParticipations?.participations, teams])

  if (!phaseId) return <ErrorPage description='Invalid course phase ID' />

  return (
    <QueryGate
      queries={[participationsQuery, teamsQuery]}
      errorFallback={({ refetch }) => (
        <ErrorPage onRetry={refetch} description='Could not fetch participants or teams' />
      )}
    >
      <div id='table-view' className='relative flex flex-col'>
        <ManagementPageHeader>Self Team Allocation Participants</ManagementPageHeader>
        <p className='text-sm text-muted-foreground mb-4'>
          This table shows all participants and their allocated teams.
        </p>
        <div className='w-full'>
          <CoursePhaseParticipationsTable
            phaseId={phaseId}
            participants={coursePhaseParticipations?.participations ?? []}
            extraColumns={extraColumns}
          />
        </div>
      </div>
    </QueryGate>
  )
}
