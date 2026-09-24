import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  type ExtraParticipantColumn,
  ManagementPageHeader,
  QueryGate,
  type TableFilter,
} from '@tumaet/prompt-ui-components'
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

  const coursePhaseConfigQuery = useGetCoursePhaseConfig()
  const coursePhaseConfig = coursePhaseConfigQuery.data
  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? false
  const { data: participations } = useGetCoursePhaseParticipations()
  const { data: scoreLevels } = useGetAllScoreLevels({ enabled: assessmentEnabled })
  const { data: teams } = useGetAllTeams()

  const assessmentCompletionsQuery = useGetAllAssessmentCompletions({ enabled: assessmentEnabled })
  const evaluationCompletionsQuery = useGetAllEvaluationCompletions()

  const assessmentCompletions = assessmentCompletionsQuery.data
  const evaluationCompletions = evaluationCompletionsQuery.data

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

  const extraFilters: TableFilter[] = [
    {
      type: 'select',
      id: 'team',
      label: 'Team',
      options: teams.map((team) => team.name),
    },
  ]

  return (
    <QueryGate
      queries={[coursePhaseConfigQuery, assessmentCompletionsQuery, evaluationCompletionsQuery]}
      errorFallback={({ refetch }) => (
        <ErrorPage message='Error loading assessments' onRetry={refetch} />
      )}
    >
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
            onClickRowAction={
              assessmentEnabled ? (row) => openAssessment(row.courseParticipationID) : undefined
            }
          />
        </div>
      </div>
    </QueryGate>
  )
}
