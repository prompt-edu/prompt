import { useQuery } from '@tanstack/react-query'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
  ErrorPage,
} from '@tumaet/prompt-ui-components'
import { AlertCircle, AlertTriangle, Loader2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import type { SurveyForm } from '../../interfaces/surveyForm'
import type { SurveyResponse } from '../../interfaces/surveyResponse'
import { getSurveyForm } from '../../network/queries/getSurveyForm'
import { getSurveyOwnResponse } from '../../network/queries/getSurveyOwnResponse'
import { SurveyFormComponent } from './components/SurveyForm'

export const StudentSurveyPage = () => {
  const { isStudentOfCourse } = useCourseStore()
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const isStudent = isStudentOfCourse(courseId ?? '')

  // Get the survey form (teams & skills)
  const {
    data: fetchedSurveyForm,
    isPending: isSurveyFormPending,
    isError: isSurveyFormError,
    refetch: refetchSurveyForm,
  } = useQuery<SurveyForm | null>({
    queryKey: ['team_allocation_survey_form', phaseId], // TODO also update on skill / teams change
    queryFn: () => getSurveyForm(phaseId ?? ''),
  })

  // Get the student's saved response, if any
  const {
    data: fetchedStudentSurveyResponse,
    isPending: isStudentSurveyResponsePending,
    isError: isStudentSurveyResponseError,
    refetch: refetchStudentSurveyResponse,
  } = useQuery<SurveyResponse>({
    queryKey: ['team_allocation_student_survey_response', phaseId],
    queryFn: () => getSurveyOwnResponse(phaseId ?? ''),
    enabled: isStudent,
  })

  const isPending = isSurveyFormPending || (isStudent && isStudentSurveyResponsePending)
  const isError = isSurveyFormError || isStudentSurveyResponseError
  const surveyStarted = fetchedSurveyForm !== null

  if (isError) {
    return (
      <ErrorPage
        onRetry={() => {
          refetchSurveyForm()
          refetchStudentSurveyResponse()
        }}
      />
    )
  }

  if (isPending) {
    return (
      <div className='container mx-auto px-4 py-16 flex flex-col items-center justify-center'>
        <Loader2 className='h-12 w-12 animate-spin text-primary mb-4' />
        <p className='text-muted-foreground animate-pulse'>Loading survey data...</p>
      </div>
    )
  }

  return (
    <div className='px-4 py-8'>
      <div className='mb-8'>
        <h1 className='text-4xl font-bold'>Team Allocation Survey</h1>
        <p className='text-muted-foreground'>
          Complete this survey to help us assign you to the most suitable team.
        </p>
      </div>

      {!isStudent && (
        <Alert variant='destructive' className='mb-8'>
          <AlertTriangle className='h-4 w-4' />
          <AlertTitle>Access Restricted</AlertTitle>
          <AlertDescription>
            The survey is disabled because you are not a student of this course.
          </AlertDescription>
        </Alert>
      )}

      {!surveyStarted && (
        <Card>
          <CardHeader className='space-y-1.5 pb-5 pt-6'>
            <div className='flex items-center gap-2.5'>
              <AlertCircle className='h-4 w-4 text-red-600' aria-hidden='true' />
              <CardTitle className='text-lg font-semibold ml-2'>Survey Not Available</CardTitle>
            </div>
            <CardDescription className='text-sm text-muted-foreground/90 mt-1 ml-9'>
              This survey hasn&apos;t started yet. Please check back later.
            </CardDescription>
          </CardHeader>
        </Card>
      )}

      {fetchedSurveyForm && surveyStarted && (
        <SurveyFormComponent
          surveyForm={fetchedSurveyForm}
          surveyResponse={fetchedStudentSurveyResponse}
          isStudent={isStudent}
        />
      )}
    </div>
  )
}
