import {
  ManagementPageHeader,
  PromptTable,
  type PromptTableColumnDef,
} from '@tumaet/prompt-ui-components'
import { type ReactNode, useMemo } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import { AssessmentDiagram } from '../components/diagrams/AssessmentDiagram'
import { ScoreLevelDistributionDiagram } from '../components/diagrams/ScoreLevelDistributionDiagram'
import { createScoreColumns } from '../components/EvaluationScore'
import { useGetAllEvaluationCompletions } from '../hooks/useGetAllEvaluationCompletions'
import { useGetAllEvaluations } from '../hooks/useGetAllEvaluations'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { getTutorLabel } from '../hooks/useTutorLabel'
import type { TutorDetailLocationState } from '../TutorEvaluationResultsPage/hooks/useTutorNavigation'
import { getScoreLevelsFromEvaluations } from '../utils/getScoreLevelsFromEvaluations'
import { getTutorRows, getUniqueTutorIds, type TutorRow } from '../utils/getTutorRows'

export const TutorOverviewPage = (): ReactNode => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const path = useLocation().pathname

  const { data: teams } = useGetAllTeams()
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const tutorLabel = getTutorLabel(coursePhaseConfig)
  const { data: evaluations } = useGetAllEvaluations()

  const { data: evaluationCompletions } = useGetAllEvaluationCompletions()

  const data: TutorRow[] = useMemo(() => getTutorRows(teams), [teams])

  const tutorParticipations = useMemo(
    () => getUniqueTutorIds(data).map((tutorId) => ({ courseParticipationID: tutorId })),
    [data],
  )

  const tutorEvaluationCompletions = useMemo(
    () => evaluationCompletions.filter((completion) => completion.type === AssessmentType.TUTOR),
    [evaluationCompletions],
  )

  const tutorScoreLevels = useMemo(
    () => getScoreLevelsFromEvaluations(evaluations, AssessmentType.TUTOR),
    [evaluations],
  )

  const columns: PromptTableColumnDef<TutorRow>[] = useMemo(
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
      ...createScoreColumns<TutorRow>(tutorScoreLevels, (row) => row.tutorId),
    ],
    [tutorScoreLevels],
  )

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>{tutorLabel.title} Overview</ManagementPageHeader>

      {coursePhaseConfig?.tutorEvaluationEnabled && (
        <p className='text-sm text-muted-foreground mb-4'>
          Click on any {tutorLabel.text} to view their evaluation results from students.
        </p>
      )}

      <div className='grid gap-6 grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 mb-6'>
        <AssessmentDiagram
          participations={tutorParticipations}
          scoreLevels={tutorScoreLevels}
          completions={tutorEvaluationCompletions}
          assessmentType={AssessmentType.TUTOR}
        />
        <ScoreLevelDistributionDiagram
          participations={tutorParticipations}
          scoreLevels={tutorScoreLevels}
          title={`${tutorLabel.title} Evaluation Distribution`}
          description={`Number of ${tutorLabel.text} evaluations per score level`}
        />
      </div>

      <PromptTable<TutorRow>
        data={data}
        columns={columns}
        onRowClick={(row, orderedRows) => {
          if (coursePhaseConfig?.tutorEvaluationEnabled) {
            navigate(`${path}/${row.tutorId}`, {
              state: {
                // A tutor of several teams has one row per team but one results page.
                orderedTutorIds: getUniqueTutorIds(orderedRows),
              } satisfies TutorDetailLocationState,
            })
          }
        }}
      />
    </div>
  )
}
