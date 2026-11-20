import { useCourseStore } from '@tumaet/prompt-shared-state'
import { Card, CardContent, QueryGate } from '@tumaet/prompt-ui-components'
import { EyeOff, Loader2 } from 'lucide-react'
import { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { useStudentAssessmentStore } from '../../../zustand/useStudentAssessmentStore'
import { AssessmentCompletion } from '../../AssessmentPage/components/AssessmentCompletion/AssessmentCompletion'
import { CategoryAssessment } from '../../AssessmentPage/components/CategoryAssessment'
import { AssessmentPrintReport } from '../../components/AssessmentPrintReport/AssessmentPrintReport'
import { useGetAllCategoriesWithCompetencies } from '../../hooks/useGetAllCategoriesWithCompetencies'
import { useGetAllTeams } from '../../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../../hooks/useGetCoursePhaseConfig'
import { useGetMyParticipation } from '../../hooks/useGetMyParticipation'
import { buildExampleAssessmentResults } from '../../utils/exampleResults'
import { hasVisibleResultSection } from '../../utils/studentResultsVisibility'
import { useGetMyAssessmentResults } from '../hooks/useGetMyAssessmentResults'

const AssessmentResultsLoader = () => (
  <div className='flex justify-center items-center h-64'>
    <Loader2 className='h-12 w-12 animate-spin text-primary' />
  </div>
)

const NoSharedResultsCard = () => (
  <Card>
    <CardContent className='flex items-start gap-3 p-6'>
      <EyeOff className='mt-0.5 h-5 w-5 shrink-0 text-muted-foreground' />
      <div className='space-y-1'>
        <h2 className='text-base font-semibold'>No results shared for this phase</h2>
        <p className='text-sm leading-6 text-muted-foreground'>
          Your instructors have not shared the assessment sheet, grade suggestions, or action items
          of this phase with students. Contact them if you have questions about your assessment.
        </p>
      </div>
    </CardContent>
  </Card>
)

interface AssessmentResultsSectionProps {
  onReadyChange?: (ready: boolean) => void
}

export const AssessmentResultsSection = ({ onReadyChange }: AssessmentResultsSectionProps) => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const resultsReleased = coursePhaseConfig?.resultsReleased ?? false
  const gradingSheetVisible = coursePhaseConfig?.gradingSheetVisible ?? false
  const actionItemsVisible = coursePhaseConfig?.actionItemsVisible ?? false
  const gradeSuggestionVisible = coursePhaseConfig?.gradeSuggestionVisible ?? false
  const hasVisibleSection = hasVisibleResultSection(coursePhaseConfig)

  const { data: myParticipation } = useGetMyParticipation({ enabled: isStudent })
  const { data: teams } = useGetAllTeams()
  const { setStudentAssessment, setAssessmentParticipation, reset } = useStudentAssessmentStore()

  const shouldFetch = isStudent && resultsReleased && hasVisibleSection
  const shouldFetchCategories = (shouldFetch || !isStudent) && gradingSheetVisible
  const resultsQuery = useGetMyAssessmentResults({ enabled: shouldFetch })
  const assessmentCategoriesQuery = useGetAllCategoriesWithCompetencies({
    enabled: shouldFetchCategories,
  })

  const { data: fetchedResults, isPending, isError } = resultsQuery
  const {
    data: assessmentCategories = [],
    isPending: isAssessmentCategoriesPending,
    isError: isAssessmentCategoriesError,
  } = assessmentCategoriesQuery

  const exampleResults = useMemo(
    () =>
      !isStudent && coursePhaseConfig
        ? buildExampleAssessmentResults(phaseId ?? '', assessmentCategories, coursePhaseConfig)
        : undefined,
    [isStudent, coursePhaseConfig, phaseId, assessmentCategories],
  )
  const results = isStudent ? fetchedResults : exampleResults

  useEffect(() => {
    if (!results) return
    setStudentAssessment({
      courseParticipationID: results.courseParticipationID,
      assessments: results.assessments,
      categoryAssessments: results.categoryAssessments ?? [],
      assessmentCompletion: results.assessmentCompletion,
      studentScore: results.studentScore,
      evaluations: [],
      independentAssessments: [],
      myIndependentAssessments: [],
    })
  }, [results, setStudentAssessment])

  // The example writes made-up results into the shared store, so clear them before another page
  // (a participant's assessment, for example) reads it, and drop the last opened participant's name.
  useEffect(() => {
    if (isStudent) return
    useStudentAssessmentStore.setState({ assessmentParticipation: undefined })
    return reset
  }, [isStudent, reset])

  const isReportReady =
    resultsReleased &&
    isStudent &&
    hasVisibleSection &&
    !isError &&
    !isAssessmentCategoriesError &&
    !isPending &&
    !(isAssessmentCategoriesPending && gradingSheetVisible) &&
    !!results

  useEffect(() => {
    onReadyChange?.(isReportReady)
  }, [isReportReady, onReadyChange])

  useEffect(() => {
    if (!myParticipation) return
    const team = teams.find((t) =>
      t.members.some((member) => member.id === myParticipation.courseParticipationID),
    )
    setAssessmentParticipation({
      ...myParticipation,
      teamID: team?.id ?? '',
    })
  }, [myParticipation, setAssessmentParticipation, teams])

  if (isStudent && !resultsReleased) return null
  if (!hasVisibleSection) return <NoSharedResultsCard />

  return (
    <QueryGate
      queries={[resultsQuery, assessmentCategoriesQuery]}
      loadingFallback={<AssessmentResultsLoader />}
    >
      {() => {
        if (!results) return <AssessmentResultsLoader />

        return (
          <>
            <div className='space-y-4 print:hidden'>
              {gradingSheetVisible &&
                assessmentCategories.map((category) => (
                  <CategoryAssessment
                    key={category.id}
                    category={category}
                    assessments={results.assessments.filter((assessment) =>
                      category.competencies
                        .map((competency) => competency.id)
                        .includes(assessment.competencyID),
                    )}
                    completed={true}
                    courseParticipationID={results.courseParticipationID}
                    peerEvaluationResults={results.peerEvaluationResults}
                    selfEvaluationResults={results.selfEvaluationResults}
                    hidePeerEvaluationDetails={true}
                  />
                ))}

              {actionItemsVisible || gradeSuggestionVisible ? (
                <AssessmentCompletion readOnly actionItems={results.actionItems} />
              ) : null}
            </div>

            {isStudent && (
              <AssessmentPrintReport
                categories={assessmentCategories}
                actionItems={results.actionItems}
              />
            )}
          </>
        )
      }}
    </QueryGate>
  )
}
