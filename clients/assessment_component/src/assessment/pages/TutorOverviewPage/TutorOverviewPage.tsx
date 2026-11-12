import {
  ManagementPageHeader,
  PromptTable,
  type PromptTableColumnDef,
  type RowAction,
  useToast,
} from '@tumaet/prompt-ui-components'
import { FileJson, FileText } from 'lucide-react'
import { type ReactNode, useMemo } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import { AssessmentDiagram } from '../components/diagrams/AssessmentDiagram'
import { ScoreLevelDistributionDiagram } from '../components/diagrams/ScoreLevelDistributionDiagram'
import { useGetAllEvaluationCompletions } from '../hooks/useGetAllEvaluationCompletions'
import { useGetAllEvaluations } from '../hooks/useGetAllEvaluations'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { getTutorLabel } from '../hooks/useTutorLabel'
import { getScoreLevelsFromEvaluations } from '../utils/getScoreLevelsFromEvaluations'
import { getTutorRows, getUniqueTutorIds, type TutorRow } from '../utils/getTutorRows'
import {
  exportTutorEvaluationsAsJson,
  exportTutorEvaluationsAsPdf,
} from './utils/exportTutorEvaluations'

export const TutorOverviewPage = (): ReactNode => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const path = useLocation().pathname

  const { data: teams } = useGetAllTeams()
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const tutorLabel = getTutorLabel(coursePhaseConfig)
  const { data: evaluations } = useGetAllEvaluations()

  const { data: evaluationCompletions } = useGetAllEvaluationCompletions()
  const tutorEvaluationEnabled = coursePhaseConfig?.tutorEvaluationEnabled ?? false
  const { toast } = useToast()

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

  const actions: RowAction<TutorRow>[] = useMemo(() => {
    if (!tutorEvaluationEnabled || !phaseId) {
      return []
    }
    const runExport = async (exportFiles: () => Promise<void>) => {
      try {
        await exportFiles()
      } catch {
        toast({
          title: 'Export failed',
          description: 'Please try again.',
          variant: 'destructive',
        })
      }
    }
    return [
      {
        label: 'Export JSON',
        icon: <FileJson className='h-4 w-4' />,
        onAction: (rows) => runExport(() => exportTutorEvaluationsAsJson(phaseId, rows)),
      },
      {
        label: 'Export PDF',
        icon: <FileText className='h-4 w-4' />,
        onAction: (rows) =>
          runExport(() => exportTutorEvaluationsAsPdf(phaseId, rows, tutorLabel.title, teams)),
      },
    ]
  }, [tutorEvaluationEnabled, phaseId, tutorLabel.title, teams, toast])

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
    ],
    [],
  )

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>{tutorLabel.title} Overview</ManagementPageHeader>

      {tutorEvaluationEnabled && (
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
        actions={actions}
        onRowClick={(row) => {
          if (tutorEvaluationEnabled) {
            navigate(`${path}/${row.tutorId}`)
          }
        }}
      />
    </div>
  )
}
