import { useCourseStore } from '@tumaet/prompt-shared-state'
import { QueryGate } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useEffect } from 'react'
import { useParams } from 'react-router-dom'

import { useStudentAssessmentStore } from '../../../zustand/useStudentAssessmentStore'
import { AssessmentCompletion } from '../../AssessmentPage/components/AssessmentCompletion/AssessmentCompletion'
import { CategoryAssessment } from '../../AssessmentPage/components/CategoryAssessment'
import { AssessmentPrintReport } from '../../components/AssessmentPrintReport/AssessmentPrintReport'
import { useGetAllCategoriesWithCompetencies } from '../../hooks/useGetAllCategoriesWithCompetencies'
import { useGetAllTeams } from '../../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../../hooks/useGetCoursePhaseConfig'
import { useGetMyParticipation } from '../../hooks/useGetMyParticipation'
import { useGetMyAssessmentResults } from '../hooks/useGetMyAssessmentResults'

const AssessmentResultsLoader = () => (
  <div className='flex justify-center items-center h-64'>
    <Loader2 className='h-12 w-12 animate-spin text-primary' />
  </div>
)

interface AssessmentResultsSectionProps {
  onReadyChange?: (ready: boolean) => void
}

export const AssessmentResultsSection = ({ onReadyChange }: AssessmentResultsSectionProps) => {
  const { courseId } = useParams<{ courseId: string }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const resultsReleased = coursePhaseConfig?.resultsReleased ?? false
  const gradingSheetVisible = coursePhaseConfig?.gradingSheetVisible ?? false
  const actionItemsVisible = coursePhaseConfig?.actionItemsVisible ?? false
  const gradeSuggestionVisible = coursePhaseConfig?.gradeSuggestionVisible ?? false
  const hasVisibleSection = gradingSheetVisible || actionItemsVisible || gradeSuggestionVisible

  const { data: myParticipation } = useGetMyParticipation({ enabled: isStudent })
  const { data: teams } = useGetAllTeams()
  const { setStudentAssessment, setAssessmentParticipation } = useStudentAssessmentStore()

  const shouldFetch = isStudent && resultsReleased
  const shouldFetchCategories = shouldFetch && gradingSheetVisible
  const resultsQuery = useGetMyAssessmentResults({ enabled: shouldFetch })
  const assessmentCategoriesQuery = useGetAllCategoriesWithCompetencies({
    enabled: shouldFetchCategories,
  })

  const { data: results, isPending, isError } = resultsQuery
  const {
    data: assessmentCategories = [],
    isPending: isAssessmentCategoriesPending,
    isError: isAssessmentCategoriesError,
  } = assessmentCategoriesQuery

  useEffect(() => {
    if (!results) return
    setStudentAssessment({
      courseParticipationID: results.courseParticipationID,
      assessments: results.assessments,
      categoryAssessments: results.categoryAssessments ?? [],
      assessmentCompletion: results.assessmentCompletion,
      studentScore: results.studentScore,
      evaluations: [],
    })
  }, [results, setStudentAssessment])

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

  if (!resultsReleased || !isStudent) return null

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

            <AssessmentPrintReport
              categories={assessmentCategories}
              actionItems={results.actionItems}
            />
          </>
        )
      }}
    </QueryGate>
  )
}
