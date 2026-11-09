import {
  ManagementPageHeader,
  PromptTable,
  type PromptTableColumnDef,
} from '@tumaet/prompt-ui-components'
import { type ReactNode, useMemo } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import { GradeSuggestionBadge, StudentScoreBadge } from '../components/badges'
import { AssessmentDiagram } from '../components/diagrams/AssessmentDiagram'
import { ScoreLevelDistributionDiagram } from '../components/diagrams/ScoreLevelDistributionDiagram'
import { useGetAllEvaluationCompletions } from '../hooks/useGetAllEvaluationCompletions'
import { useGetAllEvaluations } from '../hooks/useGetAllEvaluations'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { getTutorLabel } from '../hooks/useTutorLabel'
import type { TutorDetailLocationState } from '../TutorEvaluationResultsPage/hooks/useTutorNavigation'
import { getScoreLevelsFromEvaluations } from '../utils/getScoreLevelsFromEvaluations'

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
  const tutorLabel = getTutorLabel(coursePhaseConfig)
  const { data: evaluations } = useGetAllEvaluations()

  const { data: evaluationCompletions } = useGetAllEvaluationCompletions()

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

  const tutorScoreLevels = useMemo(
    () => getScoreLevelsFromEvaluations(evaluations, AssessmentType.TUTOR),
    [evaluations],
  )
  const tutorScoreById = useMemo(
    () => new Map(tutorScoreLevels.map((score) => [score.courseParticipationID, score])),
    [tutorScoreLevels],
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
      // The average of the students' tutor evaluations, as a level and as a grade.
      {
        id: 'scoreLevel',
        header: 'Score Level',
        accessorFn: (row) => tutorScoreById.get(row.id)?.scoreNumeric,
        cell: ({ row }) => (
          <StudentScoreBadge scoreLevel={tutorScoreById.get(row.original.id)?.scoreLevel} />
        ),
      },
      {
        id: 'grade',
        header: 'Grade',
        accessorFn: (row) => tutorScoreById.get(row.id)?.scoreNumeric,
        cell: ({ row }) => (
          <GradeSuggestionBadge
            gradeSuggestion={tutorScoreById.get(row.original.id)?.scoreNumeric}
          />
        ),
      },
    ],
    [tutorScoreById],
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
            navigate(`${path}/${row.id}`, {
              state: {
                orderedTutorIds: orderedRows.map((orderedRow) => orderedRow.id),
              } satisfies TutorDetailLocationState,
            })
          }
        }}
      />
    </div>
  )
}
