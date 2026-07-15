import { useQuery } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import type { Team } from '@tumaet/prompt-shared-state'
import { ErrorPage, ManagementPageHeader, PromptTable } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { type ReactNode, useMemo } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'
import type { EvaluationCompletion } from '../../interfaces/evaluationCompletion'
import { getAllEvaluationCompletionsInPhase } from '../../network/queries/getAllEvaluationCompletionsInPhase'
import {
  createEvaluationLookup,
  getEvaluationCounts,
} from '../AssessmentParticipantsPage/utils/evaluationUtils'
import { PeerEvaluationCompletionBadge } from '../components/badges'
import { AssessmentDiagram } from '../components/diagrams/AssessmentDiagram'
import { ScoreLevelDistributionDiagram } from '../components/diagrams/ScoreLevelDistributionDiagram'
import { useGetAllEvaluations } from '../hooks/useGetAllEvaluations'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetCoursePhaseParticipations } from '../hooks/useGetCoursePhaseParticipations'
import { getScoreLevelsFromEvaluations } from '../utils/getScoreLevelsFromEvaluations'

interface EvaluationParticipantRow {
  id: string
  firstName: string
  lastName: string
  teamName: string
  completed: number
  total: number
}

interface EvaluationParticipantsOverviewPageProps {
  assessmentType: AssessmentType.SELF | AssessmentType.PEER
}

const getTeamForParticipation = (teams: Team[], courseParticipationID: string) => {
  return teams.find((team) => team.members.some((member) => member.id === courseParticipationID))
}

const getPeerCompletionCounts = (
  courseParticipationID: string,
  teams: Team[],
  completions: EvaluationCompletion[],
) => {
  const team = getTeamForParticipation(teams, courseParticipationID)
  if (!team) {
    return { completed: 0, total: 0 }
  }

  const teamMemberIds = team.members
    .map((member) => member.id)
    .filter((id): id is string => id !== undefined && id !== courseParticipationID)

  return getEvaluationCounts(
    courseParticipationID,
    teamMemberIds,
    createEvaluationLookup(completions),
  )
}

export const EvaluationParticipantsOverviewPage = ({
  assessmentType,
}: EvaluationParticipantsOverviewPageProps): ReactNode => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const path = useLocation().pathname

  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const { data: participations } = useGetCoursePhaseParticipations()
  const { data: teams } = useGetAllTeams()
  const { data: evaluations } = useGetAllEvaluations()

  const {
    data: evaluationCompletions = [],
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['evaluationCompletions', phaseId],
    queryFn: () => getAllEvaluationCompletionsInPhase(phaseId ?? ''),
  })

  const typedCompletions = useMemo(
    () => evaluationCompletions.filter((completion) => completion.type === assessmentType),
    [assessmentType, evaluationCompletions],
  )

  const typedScoreLevels = useMemo(
    () => getScoreLevelsFromEvaluations(evaluations, assessmentType),
    [assessmentType, evaluations],
  )

  const distributionLabel = assessmentType === AssessmentType.SELF ? 'Self' : 'Peer'

  const isEnabled =
    assessmentType === AssessmentType.SELF
      ? (coursePhaseConfig?.selfEvaluationEnabled ?? false)
      : (coursePhaseConfig?.peerEvaluationEnabled ?? false)

  const pageTitle = assessmentType === AssessmentType.SELF ? 'Self Evaluations' : 'Peer Evaluations'

  const data: EvaluationParticipantRow[] = useMemo(() => {
    return participations.map((participation) => {
      const team = getTeamForParticipation(teams, participation.courseParticipationID)
      const selfCompletion = typedCompletions.find(
        (completion) =>
          completion.courseParticipationID === participation.courseParticipationID &&
          completion.authorCourseParticipationID === participation.courseParticipationID &&
          completion.completed,
      )
      const peerCounts = getPeerCompletionCounts(
        participation.courseParticipationID,
        teams,
        typedCompletions,
      )
      const counts =
        assessmentType === AssessmentType.SELF
          ? { completed: selfCompletion ? 1 : 0, total: 1 }
          : peerCounts

      return {
        id: participation.courseParticipationID,
        firstName: participation.student.firstName,
        lastName: participation.student.lastName,
        teamName: team?.name ?? '',
        completed: counts.completed,
        total: counts.total,
      }
    })
  }, [assessmentType, participations, teams, typedCompletions])

  const columns: ColumnDef<EvaluationParticipantRow>[] = useMemo(
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
      {
        id: 'completion',
        header: 'Completion',
        accessorFn: (row) => (row.total > 0 ? row.completed / row.total : 0),
        cell: ({ row }) => (
          <PeerEvaluationCompletionBadge
            completed={row.original.completed}
            total={row.original.total}
          />
        ),
        enableSorting: true,
      },
    ],
    [],
  )

  if (isError) {
    return <ErrorPage message={`Error loading ${pageTitle.toLowerCase()}`} onRetry={refetch} />
  }

  if (isPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>{pageTitle}</ManagementPageHeader>

      {isEnabled && (
        <p className='text-sm text-muted-foreground mb-4'>
          Click on a participant to view their evaluation results.
        </p>
      )}

      <div className='grid gap-6 grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 mb-6'>
        <AssessmentDiagram
          participations={participations}
          scoreLevels={typedScoreLevels}
          completions={typedCompletions}
          assessmentType={assessmentType}
        />
        <ScoreLevelDistributionDiagram
          participations={participations}
          scoreLevels={typedScoreLevels}
          title={`${distributionLabel} Evaluation Distribution`}
          description='Number of participants per score level'
        />
      </div>

      <PromptTable<EvaluationParticipantRow>
        data={data}
        columns={columns}
        onRowClick={(row) => {
          if (isEnabled) {
            navigate(`${path}/${row.id}`)
          }
        }}
      />
    </div>
  )
}
