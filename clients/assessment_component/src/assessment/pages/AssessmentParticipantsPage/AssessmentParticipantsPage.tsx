import { useMemo } from 'react'
import { useNavigate, useLocation, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'

import { ManagementPageHeader, ErrorPage, TableFilter } from '@tumaet/prompt-ui-components'
import { CoursePhaseParticipationsTable } from '@/components/pages/CoursePhaseParticipationsTable/CoursePhaseParticipationsTable'

import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'
import { useParticipationStore } from '../../zustand/useParticipationStore'
import { useScoreLevelStore } from '../../zustand/useScoreLevelStore'
import { useTeamStore } from '../../zustand/useTeamStore'

import { useGetAllAssessmentCompletions } from '../hooks/useGetAllAssessmentCompletions'
import { getAllEvaluationCompletionsInPhase } from '../../network/queries/getAllEvaluationCompletionsInPhase'

import { AssessmentType } from '../../interfaces/assessmentType'

import { AssessmentDiagram } from '../components/diagrams/AssessmentDiagram'
import { ScoreLevelDistributionDiagram } from '../components/diagrams/ScoreLevelDistributionDiagram'
import { GradeDistributionDiagram } from '../components/diagrams/GradeDistributionDiagram'

import {
  createScoreLevelColumn,
  createGradeSuggestionColumn,
  createTeamColumn,
  createSelfEvalStatusColumn,
  createPeerEvalStatusColumn,
  createTutorEvalStatusColumn,
} from './columns'
import { ExtraParticipantColumn } from '@/components/pages/CoursePhaseParticipationsTable/table/participationRow'

export const AssessmentParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const path = useLocation().pathname

  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const { participations } = useParticipationStore()
  const { scoreLevels } = useScoreLevelStore()
  const { teams } = useTeamStore()

  const {
    data: assessmentCompletions,
    isPending: isAssessmentCompletionsPending,
    isError: isAssessmentCompletionsError,
    refetch: refetchAssessmentCompletions,
  } = useGetAllAssessmentCompletions()

  const {
    data: evaluationCompletions,
    isPending: isEvaluationCompletionsPending,
    isError: isEvaluationCompletionsError,
    refetch: refetchEvaluationCompletions,
  } = useQuery({
    queryKey: ['evaluationCompletions', phaseId],
    queryFn: () => getAllEvaluationCompletionsInPhase(phaseId ?? ''),
  })

  const isError = isAssessmentCompletionsError || isEvaluationCompletionsError
  const isPending = isAssessmentCompletionsPending || isEvaluationCompletionsPending
  const refetch = () => {
    refetchAssessmentCompletions()
    refetchEvaluationCompletions()
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
    if (!scoreLevels) return []

    const columns = [
      createScoreLevelColumn(scoreLevels),
      createGradeSuggestionColumn(assessmentCompletions),
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

  if (isError) {
    return <ErrorPage message='Error loading assessments' onRetry={refetch} />
  }
  if (isPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  return (
    <div id='table-view' className='relative flex flex-col'>
      <ManagementPageHeader>Assessment Participants</ManagementPageHeader>
      <p className='text-sm text-muted-foreground mb-4'>
        Click on a participant to view/edit their assessment.
      </p>
      <div className='grid gap-6 grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 mb-6'>
        <AssessmentDiagram
          participations={participations}
          scoreLevels={scoreLevels}
          completions={assessmentCompletions}
        />
        <GradeDistributionDiagram participations={participations} grades={completedGrades} />
        <ScoreLevelDistributionDiagram participations={participations} scoreLevels={scoreLevels} />
      </div>
      <div className='w-full'>
        <CoursePhaseParticipationsTable
          phaseId={phaseId!}
          participants={participations ?? []}
          extraColumns={extraColumns}
          extraFilters={extraFilters}
          onClickRowAction={(row) => navigate(`${path}/${row.courseParticipationID}`)}
        />
      </div>
    </div>
  )
}
