import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  type ExtraParticipantColumn,
  LoadingPage,
  ManagementPageHeader,
  type ParticipantRow,
  type RowAction,
  type TableFilter,
  useToast,
} from '@tumaet/prompt-ui-components'
import { isAxiosError } from 'axios'
import { Lock, Unlock } from 'lucide-react'
import { useMemo, useRef } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import { AssessmentDiagram } from '../components/diagrams/AssessmentDiagram'
import { GradeDistributionDiagram } from '../components/diagrams/GradeDistributionDiagram'
import { ScoreLevelDistributionDiagram } from '../components/diagrams/ScoreLevelDistributionDiagram'
import { useGetAllAssessmentCompletions } from '../hooks/useGetAllAssessmentCompletions'
import { useGetAllEvaluationCompletions } from '../hooks/useGetAllEvaluationCompletions'
import { useGetAllScoreLevels } from '../hooks/useGetAllScoreLevels'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetCoursePhaseParticipations } from '../hooks/useGetCoursePhaseParticipations'
import {
  createGradeSuggestionColumn,
  createPeerEvalStatusColumn,
  createScoreLevelColumn,
  createSelfEvalStatusColumn,
  createTeamColumn,
  createTutorEvalStatusColumn,
} from './columns'
import { useMarkAssessmentsAsCompleted } from './hooks/useMarkAssessmentsAsCompleted'
import { useUnmarkAssessmentsAsCompleted } from './hooks/useUnmarkAssessmentsAsCompleted'
import {
  canMarkAnyAsCompleted,
  canUnmarkAny,
  pluralizeAssessments,
  summarizeMarkResult,
  summarizeUnmarkResult,
} from './utils/completionActions'

const toParticipationIDs = (rows: ParticipantRow[]) => rows.map((row) => row.courseParticipationID)

const serverErrorMessage = (error: unknown): string =>
  (isAxiosError<{ error?: string }>(error) ? error.response?.data?.error : undefined) ??
  'Please try again.'

export const AssessmentParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const path = useLocation().pathname
  // shared table's onClickRowAction doesn't forward the event, so capture the modifier here
  const openInNewTabRef = useRef(false)

  const openAssessment = (courseParticipationID: string) => {
    const target = `${path}/${courseParticipationID}`
    if (openInNewTabRef.current) {
      window.open(`${window.location.origin}${target}`, '_blank', 'noopener,noreferrer')
      return
    }
    navigate(target)
  }

  const {
    data: coursePhaseConfig,
    isPending: isCoursePhaseConfigPending,
    isError: isCoursePhaseConfigError,
    refetch: refetchCoursePhaseConfig,
  } = useGetCoursePhaseConfig()
  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? false
  const { data: participations } = useGetCoursePhaseParticipations()
  const { data: scoreLevels } = useGetAllScoreLevels({ enabled: assessmentEnabled })
  const { data: teams } = useGetAllTeams()

  const {
    data: assessmentCompletions,
    isPending: isAssessmentCompletionsPending,
    isError: isAssessmentCompletionsError,
    refetch: refetchAssessmentCompletions,
  } = useGetAllAssessmentCompletions({ enabled: assessmentEnabled })

  const {
    data: evaluationCompletions,
    isPending: isEvaluationCompletionsPending,
    isError: isEvaluationCompletionsError,
    refetch: refetchEvaluationCompletions,
  } = useGetAllEvaluationCompletions()

  const isError =
    isCoursePhaseConfigError ||
    (assessmentEnabled && isAssessmentCompletionsError) ||
    isEvaluationCompletionsError
  const isPending =
    isCoursePhaseConfigPending ||
    (assessmentEnabled && isAssessmentCompletionsPending) ||
    isEvaluationCompletionsPending
  const refetch = () => {
    refetchCoursePhaseConfig()
    refetchEvaluationCompletions()
    if (assessmentEnabled) {
      refetchAssessmentCompletions()
    }
  }

  const selfEvaluationCompletions = useMemo(() => {
    return (
      evaluationCompletions?.filter((evaluation) => evaluation.type === AssessmentType.SELF) ?? []
    )
  }, [evaluationCompletions])

  const peerEvaluationCompletions = useMemo(() => {
    return (
      evaluationCompletions?.filter((evaluation) => evaluation.type === AssessmentType.PEER) ?? []
    )
  }, [evaluationCompletions])

  const tutorEvaluationCompletions = useMemo(() => {
    return (
      evaluationCompletions?.filter((evaluation) => evaluation.type === AssessmentType.TUTOR) ?? []
    )
  }, [evaluationCompletions])

  const completedGrades = useMemo(() => {
    const completedGradings = assessmentCompletions?.filter((a) => a.completed) ?? []
    return completedGradings.map((completion) => completion.gradeSuggestion)
  }, [assessmentCompletions])

  const extraColumns: ExtraParticipantColumn<any>[] = useMemo(() => {
    const columns = [
      ...(assessmentEnabled
        ? [createScoreLevelColumn(scoreLevels), createGradeSuggestionColumn(assessmentCompletions)]
        : []),
      createTeamColumn(teams, participations),
      createSelfEvalStatusColumn(
        selfEvaluationCompletions,
        coursePhaseConfig?.selfEvaluationEnabled ?? false,
      ),
      createPeerEvalStatusColumn(
        peerEvaluationCompletions,
        teams,
        participations,
        coursePhaseConfig?.peerEvaluationEnabled ?? false,
      ),
      createTutorEvalStatusColumn(
        tutorEvaluationCompletions,
        teams,
        participations,
        coursePhaseConfig?.tutorEvaluationEnabled ?? false,
      ),
    ]

    return columns.filter((column): column is ExtraParticipantColumn<any> => column !== undefined)
  }, [
    participations,
    teams,
    scoreLevels,
    assessmentCompletions,
    assessmentEnabled,
    coursePhaseConfig,
    selfEvaluationCompletions,
    peerEvaluationCompletions,
    tutorEvaluationCompletions,
  ])

  const { toast } = useToast()
  const { mutateAsync: markAsCompleted } = useMarkAssessmentsAsCompleted()
  const { mutateAsync: unmarkAsCompleted } = useUnmarkAssessmentsAsCompleted()
  const deadline = coursePhaseConfig?.deadline
  const isDeadlinePassed = deadline ? new Date() > new Date(deadline) : false

  const extraActions: RowAction<ParticipantRow>[] = useMemo(() => {
    if (!assessmentEnabled) {
      return []
    }
    return [
      {
        label: 'Mark as final',
        icon: <Lock className='h-4 w-4' />,
        disabled: (rows) => !canMarkAnyAsCompleted(toParticipationIDs(rows), assessmentCompletions),
        confirm: {
          title: 'Mark as final',
          description: (count) =>
            `Mark ${pluralizeAssessments(count)} as final? Final assessments can no longer be edited. Assessments with unassessed competencies or no grade suggestion are skipped.`,
          confirmLabel: 'Mark as final',
        },
        onAction: async (rows) => {
          try {
            toast(summarizeMarkResult(await markAsCompleted(toParticipationIDs(rows))))
          } catch (error) {
            toast({
              title: 'Marking as final failed',
              description: serverErrorMessage(error),
              variant: 'destructive',
            })
          }
        },
      },
      {
        label: 'Unmark final',
        icon: <Unlock className='h-4 w-4' />,
        disabled: (rows) =>
          isDeadlinePassed || !canUnmarkAny(toParticipationIDs(rows), assessmentCompletions),
        confirm: {
          title: 'Unmark final',
          description: (count) =>
            `Reopen ${pluralizeAssessments(count)} for editing? Assessments that are not final are skipped.`,
          confirmLabel: 'Unmark final',
        },
        onAction: async (rows) => {
          try {
            toast(summarizeUnmarkResult(await unmarkAsCompleted(toParticipationIDs(rows))))
          } catch (error) {
            toast({
              title: 'Unmarking final failed',
              description: serverErrorMessage(error),
              variant: 'destructive',
            })
          }
        },
      },
    ]
  }, [
    assessmentEnabled,
    assessmentCompletions,
    isDeadlinePassed,
    markAsCompleted,
    unmarkAsCompleted,
    toast,
  ])

  const extraFilters: TableFilter[] = [
    {
      type: 'select',
      id: 'team',
      label: 'Team',
      options: teams.map((team) => team.name),
    },
  ]

  if (isError) {
    return <ErrorPage message='Error loading assessments' onRetry={refetch} />
  }
  if (isPending) {
    return <LoadingPage />
  }

  return (
    <div id='table-view' className='relative flex flex-col'>
      <ManagementPageHeader>
        {assessmentEnabled ? 'Assessment Participants' : 'Evaluation Participants'}
      </ManagementPageHeader>
      <p className='text-sm text-muted-foreground mb-4'>
        {assessmentEnabled
          ? 'Click on a participant to view/edit their assessment. Cmd/Ctrl-click to open it in a new tab.'
          : 'Assessment is disabled for this phase. This table tracks evaluation progress only.'}
      </p>
      {assessmentEnabled && (
        <div className='grid gap-6 grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 mb-6'>
          <AssessmentDiagram
            participations={participations}
            scoreLevels={scoreLevels}
            completions={assessmentCompletions}
          />
          <GradeDistributionDiagram participations={participations} grades={completedGrades} />
          <ScoreLevelDistributionDiagram
            participations={participations}
            scoreLevels={scoreLevels}
          />
        </div>
      )}
      <div
        className='w-full'
        onClickCapture={(e) => {
          openInNewTabRef.current = e.metaKey || e.ctrlKey
        }}
      >
        <CoursePhaseParticipationsTable
          phaseId={phaseId!}
          participants={participations ?? []}
          extraColumns={extraColumns}
          extraFilters={extraFilters}
          extraActions={extraActions}
          onClickRowAction={
            assessmentEnabled ? (row) => openAssessment(row.courseParticipationID) : undefined
          }
        />
      </div>
    </div>
  )
}
