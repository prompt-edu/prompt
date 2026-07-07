import { useQuery } from '@tanstack/react-query'
import { ErrorPage, LoadingPage } from '@tumaet/prompt-ui-components'
import { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'

import type { ActionItem } from '../../interfaces/actionItem'
import { getAllActionItemsForStudentInPhase } from '../../network/queries/getAllActionItemsForStudentInPhase'
import { useCategoryStore } from '../../zustand/useCategoryStore'
import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'
import { useParticipationStore } from '../../zustand/useParticipationStore'
import { useStudentAssessmentStore } from '../../zustand/useStudentAssessmentStore'
import { AssessmentPrintReport } from '../components/AssessmentPrintReport/AssessmentPrintReport'
import { AssessmentCompletion } from './components/AssessmentCompletion/AssessmentCompletion'
import { AssessmentExportMenu } from './components/AssessmentExportMenu'
import { AssessmentHeader } from './components/AssessmentHeader'
import { CategoryAssessment } from './components/CategoryAssessment'
import { FeedbackItemsPanel } from './components/FeedbackItemsPanel/FeedbackItemsPanel'
import { useGetFeedbackItemsForStudent } from './components/FeedbackItemsPanel/hooks/useGetFeedbackItemsForStudent'
import { PassStatusControls } from './components/PassStatusControls'
import { useGetStudentAssessment } from './hooks/useGetStudentAssessment'

export const AssessmentPage = () => {
  const { phaseId, courseParticipationID } = useParams<{
    phaseId: string
    courseParticipationID: string
  }>()

  const { setStudentAssessment, setAssessmentParticipation } = useStudentAssessmentStore()
  const { categories } = useCategoryStore()
  const { participations } = useParticipationStore()
  const participant = participations.find(
    (participation) => participation.courseParticipationID === courseParticipationID,
  )
  const { coursePhaseConfig } = useCoursePhaseConfigStore()

  const evaluationEnabled =
    coursePhaseConfig?.selfEvaluationEnabled || coursePhaseConfig?.peerEvaluationEnabled
  const { feedbackItems } = useGetFeedbackItemsForStudent(courseParticipationID ?? '')
  const { data: actionItems = [] } = useQuery<ActionItem[]>({
    queryKey: ['actionItems', phaseId, courseParticipationID],
    queryFn: () => getAllActionItemsForStudentInPhase(phaseId ?? '', courseParticipationID ?? ''),
    enabled: !!phaseId && !!courseParticipationID,
  })

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
        showFeedbackItems={!!evaluationEnabled}
      />
    </>
  )
}
