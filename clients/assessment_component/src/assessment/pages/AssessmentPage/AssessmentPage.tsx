import { ErrorPage, QueryGate } from '@tumaet/prompt-ui-components'
import { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { useStudentAssessmentStore } from '../../zustand/useStudentAssessmentStore'
import { AssessmentDisabledNotice } from '../components/AssessmentDisabledNotice'
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
  const coursePhaseConfigQuery = useGetCoursePhaseConfig()
  const coursePhaseConfig = coursePhaseConfigQuery.data
  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? false
  const { data: categories } = useGetAllCategoriesWithCompetencies({ enabled: assessmentEnabled })
  const { data: participations } = useGetCoursePhaseParticipations()
  const participant = participations.find(
    (participation) => participation.courseParticipationID === courseParticipationID,
  )

  const evaluationEnabled =
    coursePhaseConfig?.selfEvaluationEnabled || coursePhaseConfig?.peerEvaluationEnabled
  const { feedbackItems } = useGetFeedbackItemsForStudent(
    courseParticipationID ?? '',
    !!evaluationEnabled,
  )
  const { actionItems } = useGetActionItemsForStudent()

  const studentAssessmentQuery = useGetStudentAssessment({ enabled: assessmentEnabled })
  const {
    data: studentAssessment,
    isFetching: isStudentAssessmentFetching,
    isPlaceholderData: isPlaceholderStudentAssessmentData,
  } = studentAssessmentQuery
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

  return (
    <QueryGate queries={[coursePhaseConfigQuery, studentAssessmentQuery]}>
      {() => {
        if (!assessmentEnabled) return <AssessmentDisabledNotice title='Assessment' />

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
            />
          </>
        )
      }}
    </QueryGate>
  )
}
