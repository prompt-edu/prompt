import { useQuery } from '@tanstack/react-query'
import { type Team, useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  type ExtraParticipantColumn,
  ManagementPageHeader,
  type ParticipantRow,
  QueryGate,
  type TableFilter,
} from '@tumaet/prompt-ui-components'
import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useCanReadTeaseStudents } from '../../hooks/useCanReadTeaseStudents'
import type { Allocation } from '../../interfaces/allocation'
import type { StudentName } from '../../interfaces/studentNameUpdateRequest'
import type { TeaseStudent } from '../../interfaces/tease/student'
import { teamAllocationKeys } from '../../network/cache'
import { addStudentNamesToTeams } from '../../network/mutations/addStudentNamesToTeams'
import { getAllTeams } from '../../network/queries/getAllTeams'
import { getAllTeaseStudents } from '../../network/queries/getAllTeaseStudents'
import { getTeamAllocations } from '../../network/queries/getTeamAllocations'
import { SURVEY_STATUS_COLUMN_ID, SURVEY_STATUS_OPTIONS } from '../../utils/surveyStatus'
import { StudentDetailDialog } from './components/StudentDetailDialog'
import { createSurveyStatusColumn } from './components/SurveyStatusColumn'

export const TeamAllocationParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const canReadTeaseStudents = useCanReadTeaseStudents()
  const [selectedStudent, setSelectedStudent] = useState<TeaseStudent | null>(null)

  const participationsQuery = useGetCoursePhaseParticipants()

  const teamsQuery = useQuery<Team[]>({
    queryKey: teamAllocationKeys.teams(phaseId),
    queryFn: () => getAllTeams(phaseId ?? ''),
  })

  const teamAllocationsQuery = useQuery<Allocation[]>({
    queryKey: teamAllocationKeys.allocations(phaseId),
    queryFn: () => getTeamAllocations(phaseId ?? ''),
  })

  const teaseStudentsQuery = useQuery<TeaseStudent[]>({
    queryKey: teamAllocationKeys.tease.students(phaseId),
    queryFn: () => getAllTeaseStudents(phaseId ?? ''),
    enabled: canReadTeaseStudents,
  })

  const coursePhaseParticipations = participationsQuery.data
  const teams = teamsQuery.data
  const teamAllocations = teamAllocationsQuery.data
  const teaseStudents = canReadTeaseStudents ? teaseStudentsQuery.data : undefined

  const teaseStudentByParticipation = useMemo(
    () => new Map((teaseStudents ?? []).map((student) => [student.id, student])),
    [teaseStudents],
  )

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
      ...(teaseStudents
        ? [
            createSurveyStatusColumn(
              coursePhaseParticipations?.participations ?? [],
              teaseStudentByParticipation,
            ),
          ]
        : []),
    ]
  }, [
    coursePhaseParticipations?.participations,
    teams,
    teamAllocations,
    teaseStudents,
    teaseStudentByParticipation,
  ])

  const extraFilters = useMemo<TableFilter<ParticipantRow>[]>(
    () =>
      teaseStudents
        ? [
            {
              type: 'select',
              id: SURVEY_STATUS_COLUMN_ID,
              label: 'Survey',
              options: SURVEY_STATUS_OPTIONS,
            },
          ]
        : [],
    [teaseStudents],
  )

  const openStudentDetail = (row: ParticipantRow) => {
    const student = teaseStudentByParticipation.get(row.courseParticipationID)
    if (student) setSelectedStudent(student)
  }

  useEffect(() => {
    if (!coursePhaseParticipations?.participations?.length || !phaseId) return

    const requestPayload = {
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

    void addStudentNamesToTeams(phaseId, requestPayload).catch((error) => {
      console.error('Failed to update student names:', error)
    })
  }, [coursePhaseParticipations, phaseId])

  if (!phaseId) return <ErrorPage description='Invalid course phase ID' />

  return (
    <QueryGate
      queries={[
        participationsQuery,
        teamsQuery,
        teamAllocationsQuery,
        ...(canReadTeaseStudents ? [teaseStudentsQuery] : []),
      ]}
      errorFallback={({ refetch }) => (
        <ErrorPage onRetry={refetch} description='Could not fetch participants or teams' />
      )}
    >
      <div id='table-view' className='relative flex flex-col'>
        <ManagementPageHeader>Team Allocation Participants</ManagementPageHeader>
        <p className='text-sm text-muted-foreground mb-4'>
          {teaseStudents
            ? 'This table shows all participants, their survey status and their allocated teams. Click on a participant to see their survey answers.'
            : 'This table shows all participants and their allocated teams.'}
        </p>
        <div className='w-full'>
          <CoursePhaseParticipationsTable
            phaseId={phaseId}
            participants={coursePhaseParticipations?.participations ?? []}
            extraColumns={extraColumns}
            extraFilters={extraFilters}
            onClickRowAction={teaseStudents ? openStudentDetail : undefined}
          />
        </div>
        <StudentDetailDialog
          student={selectedStudent}
          open={selectedStudent !== null}
          onOpenChange={(open) => {
            if (!open) setSelectedStudent(null)
          }}
        />
      </div>
    </QueryGate>
  )
}
