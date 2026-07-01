import { useCourseStore } from '@tumaet/prompt-shared-state'
import { ErrorPage } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useEffect } from 'react'
import { useParams } from 'react-router-dom'

import { useCoursePhaseConfigStore } from '../../../zustand/useCoursePhaseConfigStore'
import { useMyParticipationStore } from '../../../zustand/useMyParticipationStore'
import { useStudentAssessmentStore } from '../../../zustand/useStudentAssessmentStore'
import { useTeamStore } from '../../../zustand/useTeamStore'
import { AssessmentCompletion } from '../../AssessmentPage/components/AssessmentCompletion/AssessmentCompletion'
import { CategoryAssessment } from '../../AssessmentPage/components/CategoryAssessment'
import { useGetAllCategoriesWithCompetencies } from '../../hooks/useGetAllCategoriesWithCompetencies'
import { useGetMyAssessmentResults } from '../hooks/useGetMyAssessmentResults'

export const AssessmentResultsSection = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')
  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const resultsReleased = coursePhaseConfig?.resultsReleased ?? false
  const gradingSheetVisible = coursePhaseConfig?.gradingSheetVisible ?? false

  const { myParticipation } = useMyParticipationStore()
  const { teams } = useTeamStore()
  const { setStudentAssessment, setAssessmentParticipation } = useStudentAssessmentStore()

  const shouldFetch = isStudent && resultsReleased
  const shouldFetchCategories = shouldFetch && gradingSheetVisible
  const {
    data: results,
    isPending,
    isError,
    refetch,
  } = useGetMyAssessmentResults({ enabled: shouldFetch })

  const {
    data: assessmentCategories = [],
    isPending: isAssessmentCategoriesPending,
    isError: isAssessmentCategoriesError,
    refetch: refetchAssessmentCategories,
  } = useGetAllCategoriesWithCompetencies({ enabled: shouldFetchCategories })

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
  if (isError || isAssessmentCategoriesError) {
    return (
      <ErrorPage
        onRetry={() => {
          refetch()
          refetchAssessmentCategories()
        }}
      />
    )
  }

  if (isPending || (isAssessmentCategoriesPending && gradingSheetVisible) || !results)
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )

  return (
    <div className='space-y-4'>
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

      {coursePhaseConfig?.actionItemsVisible || coursePhaseConfig?.gradeSuggestionVisible ? (
        <AssessmentCompletion readOnly actionItems={results.actionItems} />
      ) : null}
    </div>
  )
}
