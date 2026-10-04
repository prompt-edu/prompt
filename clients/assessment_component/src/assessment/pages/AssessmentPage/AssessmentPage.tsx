import {
  ErrorPage,
  QueryGate,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@tumaet/prompt-ui-components'
import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { Assessment } from '../../interfaces/assessment'
import type { CategoryWithCompetencies } from '../../interfaces/category'
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

const inCategory = (assessments: Assessment[], category: CategoryWithCompetencies) =>
  assessments.filter((assessment) =>
    category.competencies.some((competency) => competency.id === assessment.competencyID),
  )

export const AssessmentPage = () => {
  const { courseParticipationID } = useParams<{
    courseParticipationID: string
  }>()

  const { setStudentAssessment, setAssessmentParticipation } = useStudentAssessmentStore()
  const coursePhaseConfigQuery = useGetCoursePhaseConfig()
  const coursePhaseConfig = coursePhaseConfigQuery.data
  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? false
  const independentAssessmentEnabled = coursePhaseConfig?.independentAssessmentEnabled ?? false
  const [tab, setTab] = useState('mine')
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

        const finalAssessment = (
          <>
            {categories.map((category) => (
              <CategoryAssessment
                key={category.id}
                category={category}
                assessments={inCategory(studentAssessment.assessments, category)}
                completed={studentAssessment.assessmentCompletion.completed}
                disabled={isSwitchingParticipant}
                courseParticipationID={courseParticipationID ?? ''}
                independentAssessments={
                  independentAssessmentEnabled
                    ? studentAssessment.independentAssessments
                    : undefined
                }
              />
            ))}

            {evaluationEnabled && <FeedbackItemsPanel />}

            <AssessmentCompletion />

            <PassStatusControls
              courseParticipationID={courseParticipationID}
              disabled={isSwitchingParticipant}
            />

            <AssessmentExportMenu />
          </>
        )

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

              {independentAssessmentEnabled ? (
                <Tabs value={tab} onValueChange={setTab}>
                  <TabsList className='w-full'>
                    <TabsTrigger value='mine' className='flex-1'>
                      My assessment
                    </TabsTrigger>
                    <TabsTrigger value='final' className='flex-1'>
                      Final assessment
                    </TabsTrigger>
                  </TabsList>
                  <TabsContent value='mine' className='space-y-4'>
                    {categories.map((category) => (
                      <CategoryAssessment
                        key={category.id}
                        category={category}
                        assessments={inCategory(
                          studentAssessment.myIndependentAssessments,
                          category,
                        )}
                        completed={studentAssessment.assessmentCompletion.completed}
                        disabled={isSwitchingParticipant}
                        courseParticipationID={courseParticipationID ?? ''}
                        independent
                      />
                    ))}
                  </TabsContent>
                  <TabsContent value='final' className='space-y-4'>
                    {finalAssessment}
                  </TabsContent>
                </Tabs>
              ) : (
                finalAssessment
              )}
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
