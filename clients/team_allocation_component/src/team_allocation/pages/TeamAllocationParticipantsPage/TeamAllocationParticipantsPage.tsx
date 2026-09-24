import { useQuery } from '@tanstack/react-query'
import { type Team, useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  type ExtraParticipantColumn,
  ManagementPageHeader,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import type { Allocation } from '../../interfaces/allocation'
import type { StudentName } from '../../interfaces/studentNameUpdateRequest'
import { addStudentNamesToTeams } from '../../network/mutations/addStudentNamesToTeams'
import { getAllTeams } from '../../network/queries/getAllTeams'
import { getTeamAllocations } from '../../network/queries/getTeamAllocations'

export const TeamAllocationParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const participationsQuery = useGetCoursePhaseParticipants()

  const teamsQuery = useQuery<Team[]>({
    queryKey: ['team_allocation_team', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
  })

  const teamAllocationsQuery = useQuery<Allocation[]>({
    queryKey: ['team_allocations', phaseId],
    queryFn: () => getTeamAllocations(phaseId ?? ''),
  })

  const coursePhaseParticipations = participationsQuery.data
  const teams = teamsQuery.data
  const teamAllocations = teamAllocationsQuery.data

  const extraColumns: ExtraParticipantColumn<any>[] = useMemo(() => {
    if (!teams || !teamAllocations) return []

    const teamNameById = new Map(teams.map(({ id, name }) => [id, name]))
    const teamNameByParticipation = new Map<string, string>()

    for (const { projectId, students } of teamAllocations) {
      const teamName = teamNameById.get(projectId) ?? 'No Team'

      for (const courseParticipationID of students) {
        teamNameByParticipation.set(courseParticipationID, teamName)
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

        extraData: teamNameExtraData,
      } satisfies ExtraParticipantColumn<string>,
    ]
  }, [coursePhaseParticipations?.participations, teams, teamAllocations])

  useEffect(() => {
    if (!coursePhaseParticipations?.participations?.length || !phaseId) return

    const requestPayload = {
      coursePhaseID: phaseId,
      studentNamesPerID: coursePhaseParticipations.participations.reduce(
        (acc, p) => {
          if (p.student?.firstName && p.student?.lastName) {
            acc[p.courseParticipationID] = {
              firstName: p.student.firstName,
              lastName: p.student.lastName,
            }
          }
          return acc
        },
        {} as Record<string, StudentName>,
      ),
    }

    void addStudentNamesToTeams(requestPayload).catch((error) => {
      console.error('Failed to update student names:', error)
    })
  }, [coursePhaseParticipations, phaseId])

  if (!phaseId) return <ErrorPage description='Invalid course phase ID' />

  return (
    <QueryGate
      queries={[participationsQuery, teamsQuery, teamAllocationsQuery]}
      errorFallback={({ refetch }) => (
        <ErrorPage onRetry={refetch} description='Could not fetch participants or teams' />
      )}
    >
      <div id='table-view' className='relative flex flex-col'>
        <ManagementPageHeader>Team Allocation Participants</ManagementPageHeader>
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
