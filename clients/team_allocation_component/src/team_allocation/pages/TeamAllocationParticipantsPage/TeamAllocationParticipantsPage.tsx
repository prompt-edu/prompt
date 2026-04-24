import { ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useQuery } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import { CoursePhaseParticipationsTable } from '@tumaet/prompt-ui-components'
import { Team } from '@tumaet/prompt-shared-state'
import { getAllTeams } from '../../network/queries/getAllTeams'
import { useMemo } from 'react'
import { getTeamAllocations } from '../../network/queries/getTeamAllocations'
import { Allocation } from '../../interfaces/allocation'
import { useEffect } from 'react'
import { addStudentNamesToTeams } from '../../network/mutations/addStudentNamesToTeams'
import { StudentName } from '../../interfaces/studentNameUpdateRequest'
import { ExtraParticipantColumn } from '@tumaet/prompt-ui-components'

export const TeamAllocationParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: coursePhaseParticipations,
    isPending: isCoursePhaseParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
  } = useGetCoursePhaseParticipants()

  const {
    data: teams,
    isPending: isTeamsPending,
    isError: isTeamsError,
    refetch: refetchTeams,
  } = useQuery<Team[]>({
    queryKey: ['team_allocation_team', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
  })

  const {
    data: teamAllocations,
    isPending: isTeamAllocationsPending,
    isError: isTeamAllocationsError,
    refetch: refetchTeamAllocations,
  } = useQuery<Allocation[]>({
    queryKey: ['team_allocations', phaseId],
    queryFn: () => getTeamAllocations(phaseId ?? ''),
  })

  const extraColumns: ExtraParticipantColumn<any>[] = useMemo(() => {
    if (!teams || !teamAllocations) return []

    const teamNameById = new Map(teams.map(({ id, name }) => [id, name]))

    const getTeamNameForParticipation = (courseParticipationID: string): string => {
      const allocation = teamAllocations.find(({ students }) =>
        students.includes(courseParticipationID),
      )

      if (!allocation) return 'No Team'

      return teamNameById.get(allocation.projectId) ?? 'No Team'
    }

    const teamNameExtraData: any[] = []
    for (const { projectId, students } of teamAllocations) {
      const teamName = teamNameById.get(projectId) ?? 'No Team'

      for (const courseParticipationID of students) {
        teamNameExtraData.push({
          courseParticipationID,
          value: teamName,
          stringValue: teamName,
        })
      }
    }

    return [
      {
        id: 'allocatedTeam',
        header: 'Allocated Team',

        accessorFn: (row) => getTeamNameForParticipation(row.courseParticipationID),

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
  }, [teams, teamAllocations])

  const refetch = () => {
    refetchCoursePhaseParticipations()
    refetchTeams()
    refetchTeamAllocations()
  }

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

  const isError = isParticipationsError || isTeamsError || isTeamAllocationsError
  const isPending = isCoursePhaseParticipationsPending || isTeamsPending || isTeamAllocationsPending

  if (!phaseId) return <ErrorPage onRetry={refetch} description='Invalid course phase ID' />
  if (isError)
    return <ErrorPage onRetry={refetch} description='Could not fetch participants or teams' />
  if (isPending)
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )

  return (
    <div id='table-view' className='relative flex flex-col'>
      <ManagementPageHeader>Team Allocation Participants</ManagementPageHeader>
      <p className='text-sm text-muted-foreground mb-4'>
        This table shows all participants and their allocated teams.
      </p>
      <div className='w-full'>
        <CoursePhaseParticipationsTable
          phaseId={phaseId}
          participants={coursePhaseParticipations.participations ?? []}
          extraColumns={extraColumns}
        />
      </div>
    </div>
  )
}
