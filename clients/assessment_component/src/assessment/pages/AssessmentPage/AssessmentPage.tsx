import { ErrorPage, LoadingPage } from '@tumaet/prompt-ui-components'
import { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { useStudentAssessmentStore } from '../../zustand/useStudentAssessmentStore'
import { AssessmentPrintReport } from '../components/AssessmentPrintReport/AssessmentPrintReport'
import { useGetActionItemsForStudent } from '../hooks/useGetActionItemsForStudent'
import { useGetAllCategoriesWithCompetencies } from '../hooks/useGetAllCategoriesWithCompetencies'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetCoursePhaseParticipations } from '../hooks/useGetCoursePhaseParticipations'
import { AssessmentCompletion } from './components/AssessmentCompletion/AssessmentCompletion'
import { AssessmentExportMenu } from './components/AssessmentExportMenu'
import { AssessmentHeader } from './components/AssessmentHeader'
import { CategoryAssessment } from './components/CategoryAssessment'
import { FeedbackItemsPanel } from './components/FeedbackItemsPanel/FeedbackItemsPanel'
import { useGetFeedbackItemsForStudent } from './components/FeedbackItemsPanel/hooks/useGetFeedbackItemsForStudent'
import { PassStatusControls } from './components/PassStatusControls'
import { useGetStudentAssessment } from './hooks/useGetStudentAssessment'

export const AssessmentPage = () => {
  const { courseParticipationID } = useParams<{
    courseParticipationID: string
  }>()

  const { setStudentAssessment, setAssessmentParticipation } = useStudentAssessmentStore()
  const { data: categories } = useGetAllCategoriesWithCompetencies()
  const { data: participations } = useGetCoursePhaseParticipations()
  const participant = participations.find(
    (participation) => participation.courseParticipationID === courseParticipationID,
  )
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()

  const evaluationEnabled =
    coursePhaseConfig?.selfEvaluationEnabled || coursePhaseConfig?.peerEvaluationEnabled
  const { feedbackItems } = useGetFeedbackItemsForStudent(
    courseParticipationID ?? '',
    !!evaluationEnabled,
  )
  const { actionItems } = useGetActionItemsForStudent()

  const {
    data: studentAssessment,
    isPending: isStudentAssessmentPending,
    isFetching: isStudentAssessmentFetching,
    isPlaceholderData: isPlaceholderStudentAssessmentData,
    isError: isStudentAssessmentError,
    refetch: refetchStudentAssessment,
  } = useGetStudentAssessment()
  const isSwitchingParticipant = isStudentAssessmentFetching && isPlaceholderStudentAssessmentData

  const remainingAssessments = useMemo(() => {
    return (
      categories.reduce((acc, category) => {
        return acc + category.competencies.length
      }, 0) - (studentAssessment?.assessments?.length ?? 0)
    )
  }, [categories, studentAssessment?.assessments?.length])

  useEffect(() => {
    if (studentAssessment) {
      setStudentAssessment(studentAssessment)
    }
  }, [studentAssessment, setStudentAssessment])

  useEffect(() => {
    if (participant) {
      setAssessmentParticipation(participant)
    }
  }, [participant, setAssessmentParticipation])

  if (isStudentAssessmentError) return <ErrorPage onRetry={refetchStudentAssessment} />
  if (isStudentAssessmentPending) return <LoadingPage />

  if (!studentAssessment) {
    return (
      <ErrorPage
        title='No participant found for this course participation ID'
        description='We like what you are doing. To contribute, checkout https://github.com/prompt-edu/prompt'
      />
    )
  }

  return (
    <>
      <div className='space-y-4 print:hidden' aria-busy={isSwitchingParticipant}>
        {participant && (
          <AssessmentHeader
            participant={participant}
            studentAssessment={studentAssessment}
            remainingAssessments={remainingAssessments}
          />
        )}

        {categories.map((category) => (
          <CategoryAssessment
            key={category.id}
            category={category}
            assessments={studentAssessment.assessments.filter((assessment) =>
              category.competencies
                .map((competency) => competency.id)
                .includes(assessment.competencyID),
            )}
            completed={studentAssessment.assessmentCompletion.completed}
            disabled={isSwitchingParticipant}
            courseParticipationID={courseParticipationID ?? ''}
          />
        ))}

        {evaluationEnabled && <FeedbackItemsPanel />}

        <AssessmentCompletion />

        <PassStatusControls
          courseParticipationID={courseParticipationID}
          disabled={isSwitchingParticipant}
        />

        <AssessmentExportMenu />
      </div>

      <AssessmentPrintReport
        categories={categories}
        feedbackItems={feedbackItems}
        actionItems={actionItems}
        showFeedbackItems={evaluationEnabled}
      />
    </>
  )
}
