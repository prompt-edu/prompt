import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Button,
  Card,
  CardContent,
  ManagementPageHeader,
} from '@tumaet/prompt-ui-components'
import { ArrowLeft, Info, Printer } from 'lucide-react'
import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { printPage } from '../utils/printPage'

import { AssessmentResultsSection } from './components/AssessmentResultsSection'
import { EvaluationResultsSection } from './components/EvaluationResultsSection'

export const EvaluationResultsPage = () => {
  const navigate = useNavigate()
  const { courseId } = useParams<{ courseId: string }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const resultsReleased = coursePhaseConfig?.resultsReleased ?? false
  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? true
  const [reportReady, setReportReady] = useState(false)

  return (
    <div className='w-full px-4 py-6 text-left'>
      <div className='mb-4 print:hidden'>
        <Button variant='outline' onClick={() => navigate('..')} className='gap-2'>
          <ArrowLeft className='h-4 w-4' />
          Back to overview
        </Button>
      </div>
      <div className='print:hidden'>
        <ManagementPageHeader>
          {assessmentEnabled ? 'Assessment Results' : 'Evaluation Results'}
        </ManagementPageHeader>
      </div>

      {!isStudent && (
        <Alert className='mb-4 print:hidden'>
          <Info className='h-4 w-4' />
          <AlertTitle>Example data</AlertTitle>
          <AlertDescription>
            You are not a student of this course, so this page shows example results. Students see
            their own results in this layout, limited to the sections enabled in the settings
            {resultsReleased ? '.' : ', once you release the results.'}
          </AlertDescription>
        </Alert>
      )}

      {resultsReleased || !isStudent ? (
        assessmentEnabled ? (
          <AssessmentResultsSection onReadyChange={setReportReady} />
        ) : (
          <EvaluationResultsSection onReadyChange={setReportReady} />
        )
      ) : (
        <Card className='border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xs'>
          <CardContent className='p-6'>
            <p className='text-gray-600 dark:text-gray-300 leading-relaxed'>
              {assessmentEnabled ? 'Assessment results' : 'Evaluation results'} have not been
              released yet.
            </p>
          </CardContent>
        </Card>
      )}

      {resultsReleased && reportReady && (
        <div className='flex justify-end pt-4 print:hidden'>
          <Button variant='outline' onClick={printPage} className='gap-2'>
            <Printer className='h-4 w-4' />
            PDF / Print
          </Button>
        </div>
      )}
    </div>
  )
}
