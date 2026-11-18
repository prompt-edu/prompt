import { mapNumberToScoreLevel, type ScoreLevel, useCourseStore } from '@tumaet/prompt-shared-state'
import { Card, CardContent, QueryGate } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../../interfaces/category'
import type { StudentEvaluationResults } from '../../../interfaces/evaluationResults'
import { buildResultCategories, ResultCategoryList } from '../../components/ResultCategoryList'
import { useGetCoursePhaseConfig } from '../../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../../hooks/useGetEvaluationCategoriesWithCompetencies'
import { buildExampleEvaluationResults } from '../../utils/exampleResults'
import { useGetMyEvaluationResults } from '../hooks/useGetMyEvaluationResults'
import { EvaluationPrintReport, type EvaluationPrintSection } from './EvaluationPrintReport'

interface EvaluationResultsSectionProps {
  onReadyChange?: (ready: boolean) => void
}

const buildPrintSection = (
  categories: CategoryWithCompetencies[],
  levelByCompetency: Map<string, ScoreLevel>,
  results: StudentEvaluationResults | null,
): EvaluationPrintSection => ({
  categories: categories
    .map((category) => ({
      ...category,
      competencies: category.competencies.filter((competency) =>
        levelByCompetency.has(competency.id),
      ),
    }))
    .filter((category) => category.competencies.length > 0),
  scores: [...levelByCompetency].map(([competencyID, scoreLevel]) => ({
    id: competencyID,
    courseParticipationID: results?.courseParticipationID ?? '',
    coursePhaseID: results?.coursePhaseID ?? '',
    competencyID,
    scoreLevel,
  })),
})

const EvaluationResultsLoader = () => (
  <div className='flex h-64 items-center justify-center'>
    <Loader2 className='h-12 w-12 animate-spin text-primary' />
  </div>
)

export const EvaluationResultsSection = ({ onReadyChange }: EvaluationResultsSectionProps) => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const resultsReleased = coursePhaseConfig?.resultsReleased ?? false
  const selfEvaluationEnabled = coursePhaseConfig?.selfEvaluationEnabled ?? false
  const peerEvaluationEnabled = coursePhaseConfig?.peerEvaluationEnabled ?? false

  const shouldFetch = isStudent && resultsReleased
  const resultsQuery = useGetMyEvaluationResults({
    enabled: shouldFetch,
  })
  const { data: fetchedResults, isPending, isError } = resultsQuery

  const shouldFetchCategories = shouldFetch || !isStudent
  const selfCategoriesQuery = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.SELF,
    shouldFetchCategories && selfEvaluationEnabled,
  )
  const peerCategoriesQuery = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.PEER,
    shouldFetchCategories && peerEvaluationEnabled,
  )
  const selfCategories = selfCategoriesQuery.data
  const peerCategories = peerCategoriesQuery.data

  const exampleResults = useMemo(
    () => buildExampleEvaluationResults(phaseId ?? '', selfCategories, peerCategories),
    [phaseId, selfCategories, peerCategories],
  )
  const results = isStudent ? fetchedResults : exampleResults

  const selfLevels = useMemo(
    () => new Map((results?.selfResults ?? []).map((r) => [r.competencyID, r.scoreLevel])),
    [results],
  )

  const peerLevels = useMemo(
    () =>
      new Map(
        (results?.peerResults ?? []).map((r) => [
          r.competencyID,
          mapNumberToScoreLevel(r.averageScoreNumeric),
        ]),
      ),
    [results],
  )

  const selfSections = useMemo(
    () => buildResultCategories(selfCategories, selfLevels),
    [selfCategories, selfLevels],
  )

  const peerSections = useMemo(
    () => buildResultCategories(peerCategories, peerLevels),
    [peerCategories, peerLevels],
  )

  const selfPrintSection = useMemo(
    () => buildPrintSection(selfCategories, selfLevels, results),
    [selfCategories, selfLevels, results],
  )

  const peerPrintSection = useMemo(
    () => buildPrintSection(peerCategories, peerLevels, results),
    [peerCategories, peerLevels, results],
  )

  const hasContent = selfSections.length > 0 || peerSections.length > 0
  const isReportReady = shouldFetch && !isPending && !isError && hasContent

  useEffect(() => {
    onReadyChange?.(isReportReady)
  }, [isReportReady, onReadyChange])

  if (isStudent && !resultsReleased) return null
  return (
    <QueryGate
      queries={[resultsQuery, selfCategoriesQuery, peerCategoriesQuery]}
      loadingFallback={<EvaluationResultsLoader />}
    >
      {() => {
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
                      Averaged across your teammates. Competencies rated by only one peer are not
                      shown.
                    </p>
                  </div>
                  <ResultCategoryList categories={peerSections} />
                </div>
              )}
            </div>

            <EvaluationPrintReport self={selfPrintSection} peer={peerPrintSection} />
          </>
        )
      }}
    </QueryGate>
  )
}
