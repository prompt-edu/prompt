import { mapNumberToScoreLevel, type ScoreLevel, useCourseStore } from '@tumaet/prompt-shared-state'
import { Card, CardContent, ErrorPage } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../../interfaces/category'
import { ScoreChip } from '../../components/PrintReport/ScoreChip'
import { useGetCoursePhaseConfig } from '../../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../../hooks/useGetEvaluationCategoriesWithCompetencies'
import { getScoreLevelDescription } from '../../utils/getScoreLevelDescription'
import { useGetMyEvaluationResults } from '../hooks/useGetMyEvaluationResults'
import { EvaluationPrintReport } from './EvaluationPrintReport'

interface EvaluationResultsSectionProps {
  onReadyChange?: (ready: boolean) => void
}

interface EvaluationResultRow {
  competencyID: string
  name: string
  description: string
  scoreLevel: ScoreLevel
}

export interface EvaluationResultCategory {
  id: string
  name: string
  rows: EvaluationResultRow[]
}

const buildCategories = (
  categories: CategoryWithCompetencies[],
  levelByCompetency: Map<string, ScoreLevel>,
): EvaluationResultCategory[] =>
  categories
    .map((category) => ({
      id: category.id,
      name: category.name,
      rows: category.competencies.flatMap((competency) => {
        const scoreLevel = levelByCompetency.get(competency.id)
        if (!scoreLevel) return []
        return [
          {
            competencyID: competency.id,
            name: competency.name,
            description: getScoreLevelDescription(scoreLevel, competency),
            scoreLevel,
          },
        ]
      }),
    }))
    .filter((category) => category.rows.length > 0)

const ResultCategoryList = ({ categories }: { categories: EvaluationResultCategory[] }) => (
  <div className='space-y-4'>
    {categories.map((category) => (
      <Card key={category.id}>
        <CardContent className='space-y-3 p-4'>
          <h3 className='text-sm font-semibold'>{category.name}</h3>
          {category.rows.map((row) => (
            <div
              key={row.competencyID}
              className='flex flex-wrap items-start justify-between gap-2 border-b border-border pb-2 last:border-0 last:pb-0'
            >
              <div className='min-w-0 flex-1'>
                <p className='text-sm font-medium'>{row.name}</p>
                <p className='text-sm text-muted-foreground'>{row.description}</p>
              </div>
              <ScoreChip scoreLevel={row.scoreLevel} />
            </div>
          ))}
        </CardContent>
      </Card>
    ))}
  </div>
)

export const EvaluationResultsSection = ({ onReadyChange }: EvaluationResultsSectionProps) => {
  const { courseId } = useParams<{ courseId: string }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const resultsReleased = coursePhaseConfig?.resultsReleased ?? false
  const selfEvaluationEnabled = coursePhaseConfig?.selfEvaluationEnabled ?? false
  const peerEvaluationEnabled = coursePhaseConfig?.peerEvaluationEnabled ?? false

  const shouldFetch = isStudent && resultsReleased
  const {
    data: results,
    isPending,
    isError,
    refetch,
  } = useGetMyEvaluationResults({
    enabled: shouldFetch,
  })

  const { data: selfCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.SELF,
    shouldFetch && selfEvaluationEnabled,
  )
  const { data: peerCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.PEER,
    shouldFetch && peerEvaluationEnabled,
  )

  const selfSections = useMemo(
    () =>
      buildCategories(
        selfCategories,
        new Map((results?.selfResults ?? []).map((r) => [r.competencyID, r.scoreLevel])),
      ),
    [selfCategories, results],
  )

  const peerSections = useMemo(
    () =>
      buildCategories(
        peerCategories,
        new Map(
          (results?.peerResults ?? []).map((r) => [
            r.competencyID,
            mapNumberToScoreLevel(r.averageScoreNumeric),
          ]),
        ),
      ),
    [peerCategories, results],
  )

  const hasContent = selfSections.length > 0 || peerSections.length > 0
  const isReportReady = shouldFetch && !isPending && !isError && hasContent

  useEffect(() => {
    onReadyChange?.(isReportReady)
  }, [isReportReady, onReadyChange])

  if (!resultsReleased || !isStudent) return null
  if (isError) return <ErrorPage onRetry={refetch} />
  if (isPending) {
    return (
      <div className='flex h-64 items-center justify-center'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  if (!hasContent) {
    return (
      <Card>
        <CardContent className='p-6'>
          <p className='text-sm text-muted-foreground'>
            No evaluation results are available for you in this phase yet.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <>
      <div className='space-y-6 print:hidden'>
        {selfSections.length > 0 && (
          <div className='space-y-3'>
            <h2 className='text-lg font-semibold'>Your self-evaluation</h2>
            <ResultCategoryList categories={selfSections} />
          </div>
        )}

        {peerSections.length > 0 && (
          <div className='space-y-3'>
            <div className='space-y-1'>
              <h2 className='text-lg font-semibold'>Peer feedback</h2>
              <p className='text-sm text-muted-foreground'>
                Averaged across your teammates. Competencies rated by only one peer are not shown.
              </p>
            </div>
            <ResultCategoryList categories={peerSections} />
          </div>
        )}
      </div>

      <EvaluationPrintReport selfSections={selfSections} peerSections={peerSections} />
    </>
  )
}
