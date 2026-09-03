import { useQuery } from '@tanstack/react-query'
import type { Team } from '@tumaet/prompt-shared-state'
import {
  Button,
  ErrorPage,
  getStudentName,
  LoadingPage,
  ManagementPageHeader,
  PromptTable,
  type PromptTableColumnDef,
} from '@tumaet/prompt-ui-components'
import { Loader2, Printer } from 'lucide-react'
import { type ReactNode, useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import type { EvaluationCompletion } from '../../interfaces/evaluationCompletion'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'
import {
  createEvaluationLookup,
  getEvaluationCounts,
} from '../AssessmentParticipantsPage/utils/evaluationUtils'
import { PeerEvaluationCompletionBadge } from '../components/badges'
import { AssessmentDiagram } from '../components/diagrams/AssessmentDiagram'
import { ScoreLevelDistributionDiagram } from '../components/diagrams/ScoreLevelDistributionDiagram'
import { PrintReport } from '../components/PrintReport/PrintReport'
import { useGetAllEvaluationCompletions } from '../hooks/useGetAllEvaluationCompletions'
import { useGetAllEvaluations } from '../hooks/useGetAllEvaluations'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetCoursePhaseParticipations } from '../hooks/useGetCoursePhaseParticipations'
import { useGetEvaluationCategoriesWithCompetencies } from '../hooks/useGetEvaluationCategoriesWithCompetencies'
import { getScoreLevelsFromEvaluations } from '../utils/getScoreLevelsFromEvaluations'
import { getTeamMemberName } from '../utils/getTeamMemberName'
import { printPage } from '../utils/printPage'

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

  const isEnabled =
    assessmentType === AssessmentType.SELF
      ? (coursePhaseConfig?.selfEvaluationEnabled ?? false)
      : (coursePhaseConfig?.peerEvaluationEnabled ?? false)

  const { data: categories } = useGetEvaluationCategoriesWithCompetencies(assessmentType, isEnabled)

  const {
    data: evaluationCompletions = [],
    isPending,
    isError,
    refetch,
  } = useGetAllEvaluationCompletions()

  // Feedback items are only needed for the bulk report, so they are fetched on
  // demand. A counter rather than a boolean: React Query reports isSuccess for
  // cached data even while disabled, and a latched boolean would ignore the
  // second click.
  const [printRequests, setPrintRequests] = useState(0)
  const { data: allFeedbackItems = [], isSuccess: feedbackReady } = useQuery({
    queryKey: assessmentKeys.feedbackItems.inPhase(phaseId),
    queryFn: () => assessmentApi.feedbackItems.listInPhase(phaseId ?? ''),
    enabled: printRequests > 0,
  })

  const reportsReady = printRequests > 0 && feedbackReady

  useEffect(() => {
    if (printRequests > 0 && feedbackReady) printPage()
  }, [printRequests, feedbackReady])

  const typedCompletions = useMemo(
    () => evaluationCompletions.filter((completion) => completion.type === assessmentType),
    [assessmentType, evaluationCompletions],
  )

  const typedScoreLevels = useMemo(
    () => getScoreLevelsFromEvaluations(evaluations, assessmentType),
    [assessmentType, evaluations],
  )

  const distributionLabel = assessmentType === AssessmentType.SELF ? 'Self' : 'Peer'

  const pageTitle = assessmentType === AssessmentType.SELF ? 'Self Evaluations' : 'Peer Evaluations'
  const reportTitle = `${distributionLabel} Evaluation Results`

  const bulkReports = useMemo(() => {
    if (!reportsReady) return []
    return [...participations]
      .sort((a, b) => a.student.lastName.localeCompare(b.student.lastName))
      .map((participation) => ({
        courseParticipationID: participation.courseParticipationID,
        studentName: getStudentName(participation.student),
        teamName: getTeamForParticipation(teams, participation.courseParticipationID)?.name,
        scores: evaluations
          .filter(
            (evaluation) =>
              evaluation.type === assessmentType &&
              evaluation.courseParticipationID === participation.courseParticipationID,
          )
          .map((evaluation) => ({
            ...evaluation,
            authorName: getTeamMemberName(teams, evaluation.authorCourseParticipationID),
          })),
        feedbackItems: allFeedbackItems.filter(
          (item) =>
            item.type === assessmentType &&
            item.courseParticipationID === participation.courseParticipationID,
        ),
      }))
      .filter((report) => report.scores.length > 0)
  }, [allFeedbackItems, assessmentType, evaluations, participations, reportsReady, teams])

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

  const columns: PromptTableColumnDef<EvaluationParticipantRow>[] = useMemo(
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
    return <LoadingPage />
  }

  return (
    <>
      <div className='space-y-4 print:hidden'>
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

        {isEnabled && categories.length > 0 && (
          <div className='flex justify-end pt-4'>
            <Button
              variant='outline'
              className='gap-2'
              disabled={printRequests > 0 && !feedbackReady}
              onClick={() => setPrintRequests((requests) => requests + 1)}
            >
              {printRequests > 0 && !feedbackReady ? (
                <Loader2 className='h-4 w-4 animate-spin' />
              ) : (
                <Printer className='h-4 w-4' />
              )}
              PDF / Print All
            </Button>
          </div>
        )}
      </div>

      {bulkReports.map((report, index) => (
        <PrintReport
          key={report.courseParticipationID}
          className={index > 0 ? 'break-before-page' : undefined}
          title={`${reportTitle} for ${report.studentName}`}
          subtitle={report.teamName}
          categories={categories}
          scores={report.scores}
          feedbackItems={report.feedbackItems}
        />
      ))}
    </>
  )
}
