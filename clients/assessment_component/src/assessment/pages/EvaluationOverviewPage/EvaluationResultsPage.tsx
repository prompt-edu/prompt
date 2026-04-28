import { useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { ManagementPageHeader, Card, CardContent, Button } from '@tumaet/prompt-ui-components'

import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'

import { AssessmentResultsSection } from './components/AssessmentResultsSection'

export const EvaluationResultsPage = () => {
  const navigate = useNavigate()
  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const resultsReleased = coursePhaseConfig?.resultsReleased ?? false

  return (
    <div className='w-full px-4 py-6 text-left'>
      <div className='mb-4'>
        <Button variant='outline' onClick={() => navigate('..')} className='gap-2'>
          <ArrowLeft className='h-4 w-4' />
          Back to overview
        </Button>
      </div>
      <ManagementPageHeader>Assessment Results</ManagementPageHeader>

      {resultsReleased ? (
        <AssessmentResultsSection />
      ) : (
        <Card className='border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xs'>
          <CardContent className='p-6'>
            <p className='text-gray-600 dark:text-gray-300 leading-relaxed'>
              Assessment results have not been released yet.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
