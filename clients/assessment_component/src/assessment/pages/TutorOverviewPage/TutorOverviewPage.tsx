import { useQuery } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { ManagementPageHeader, PromptTable } from '@tumaet/prompt-ui-components'
import { type ReactNode, useMemo } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import { getAllEvaluationCompletionsInPhase } from '../../network/queries/getAllEvaluationCompletionsInPhase'
import { AssessmentDiagram } from '../components/diagrams/AssessmentDiagram'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'

interface TutorRow {
  id: string
  firstName: string
  lastName: string
  teamName: string
}

export const TutorOverviewPage = (): ReactNode => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const path = useLocation().pathname

  const { data: teams } = useGetAllTeams()
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()

  const { data: evaluationCompletions = [] } = useQuery({
    queryKey: ['evaluationCompletions', phaseId],
    queryFn: () => getAllEvaluationCompletionsInPhase(phaseId ?? ''),
  })

  const data: TutorRow[] = useMemo(() => {
    return teams.flatMap((team) =>
      team.tutors
        .filter((tutor) => tutor.id)
        .map((tutor) => ({
          id: tutor.id!,
          firstName: tutor.firstName,
          lastName: tutor.lastName,
          teamName: team.name,
        })),
    )
  }, [teams])

  const tutorParticipations = useMemo(
    () => data.map((tutor) => ({ courseParticipationID: tutor.id })),
    [data],
  )

  const tutorEvaluationCompletions = useMemo(
    () => evaluationCompletions.filter((completion) => completion.type === AssessmentType.TUTOR),
    [evaluationCompletions],
  )

  const columns: ColumnDef<TutorRow>[] = useMemo(
    () => [
      {
        accessorKey: 'firstName',
        header: 'First Name',
      },
      {
        accessorKey: 'lastName',
        header: 'Last Name',
      },
      {
        accessorKey: 'teamName',
        header: 'Team',
      },
    ],
    [],
  )

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>Tutor Overview</ManagementPageHeader>

      {coursePhaseConfig?.tutorEvaluationEnabled && (
        <p className='text-sm text-muted-foreground mb-4'>
          Click on a tutor to view their evaluation results from students.
        </p>
      )}

      <div className='grid gap-6 grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 mb-6'>
        <AssessmentDiagram
          participations={tutorParticipations}
          scoreLevels={[]}
          completions={tutorEvaluationCompletions}
          assessmentType={AssessmentType.TUTOR}
        />
      </div>

      <PromptTable<TutorRow>
        data={data}
        columns={columns}
        onRowClick={(row) => {
          if (coursePhaseConfig?.tutorEvaluationEnabled) {
            navigate(`${path}/${row.id}`)
          }
        }}
      />
    </div>
  )
}
