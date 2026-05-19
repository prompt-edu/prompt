import { AlertCircle, CheckCircle2, Calendar } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'
import { Card, CardContent, Button } from '@tumaet/prompt-ui-components'
import { format } from 'date-fns'

import { useCoursePhaseConfigStore } from '../../../zustand/useCoursePhaseConfigStore'

interface EvaluationInfoHeaderProps {
  allEvaluationsCompleted: boolean
  resultsLink?: string
}

export const EvaluationInfoHeader = ({
  allEvaluationsCompleted,
  resultsLink,
}: EvaluationInfoHeaderProps) => {
  const navigate = useNavigate()
  const { coursePhaseConfig } = useCoursePhaseConfigStore()

  const now = new Date()
  const selfEvaluationStarted =
    !coursePhaseConfig?.selfEvaluationStart ||
    now >= new Date(coursePhaseConfig.selfEvaluationStart)
  const peerEvaluationStarted =
    !coursePhaseConfig?.peerEvaluationStart ||
    now >= new Date(coursePhaseConfig.peerEvaluationStart)

  const evaluationsNotStarted =
    !selfEvaluationStarted || (coursePhaseConfig?.peerEvaluationEnabled && !peerEvaluationStarted)

  const resultsReleased = coursePhaseConfig?.resultsReleased ?? false

  const gradeSuggestionVisible = coursePhaseConfig?.gradeSuggestionVisible ?? true
  const actionItemsVisible = coursePhaseConfig?.actionItemsVisible ?? true
  const bothVisible = gradeSuggestionVisible && actionItemsVisible

  const getResultsAvailableText = () => {
    if (bothVisible) {
      return (
        'Assessment results have been released and are now available. ' +
        'You can view your action items and grade suggestions below to understand areas for improvement.'
      )
    } else if (gradeSuggestionVisible) {
      return (
        'Assessment results have been released and are now available. ' +
        'You can view your grade suggestion below to understand your performance.'
      )
    } else if (actionItemsVisible) {
      return (
        'Assessment results have been released and are now available. ' +
        'You can view your action items below to understand areas for improvement.'
      )
    }
    return (
      'Assessment results have been released and are now available. ' +
      'You will receive further information about your assessment from your tutor / instructor.'
    )
  }

  return (
    <Card className='mb-8 border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xs'>
      <CardContent className='p-6'>
        <div className='flex items-start gap-4'>
          <div className='shrink-0'>
            {resultsReleased ? (
              <CheckCircle2 className='h-8 w-8 text-blue-500 dark:text-blue-400' />
            ) : allEvaluationsCompleted ? (
              <CheckCircle2 className='h-8 w-8 text-green-500 dark:text-green-400' />
            ) : evaluationsNotStarted ? (
              <Calendar className='h-8 w-8 text-orange-500 dark:text-orange-400' />
            ) : (
              <AlertCircle className='h-8 w-8 text-blue-500 dark:text-blue-400' />
            )}
          </div>
          <div className='flex-1'>
            <h2 className='text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2'>
              {resultsReleased
                ? 'Assessment Results Released'
                : allEvaluationsCompleted
                  ? 'All Evaluations Completed!'
                  : evaluationsNotStarted
                    ? 'Evaluations Not Yet Available'
                    : 'Instructions'}
            </h2>
            {resultsReleased ? (
              <>
                <p className='text-gray-600 dark:text-gray-300 leading-relaxed'>
                  {getResultsAvailableText()}
                </p>
                {resultsLink && (
                  <div className='mt-3'>
                    <Button
                      variant='outline'
                      onClick={() => navigate(resultsLink)}
                      className='gap-2'
                    >
                      Go to your results
                      <ArrowRight className='h-4 w-4' />
                    </Button>
                  </div>
                )}
              </>
            ) : allEvaluationsCompleted ? (
              <p className='text-gray-600 dark:text-gray-300 leading-relaxed'>
                Congratulations! You have successfully completed all required evaluations. Your
                self-evaluation
                {coursePhaseConfig?.peerEvaluationEnabled && ' and peer evaluations'} have been
                submitted and will be reviewed accordingly.
              </p>
            ) : evaluationsNotStarted ? (
              <div className='space-y-3'>
                <p className='text-gray-600 dark:text-gray-300 leading-relaxed'>
                  {!selfEvaluationStarted && !peerEvaluationStarted
                    ? 'Self and peer evaluations will be available soon. Please check back after the start date.'
                    : !selfEvaluationStarted
                      ? 'Self evaluation will be available soon. Please check back after the start date.'
                      : 'Peer evaluation will be available soon. Please check back after the start date.'}
                </p>
                <div className='flex flex-wrap gap-4 text-sm text-gray-500 dark:text-gray-400'>
                  {!selfEvaluationStarted && coursePhaseConfig?.selfEvaluationStart && (
                    <div className='flex items-center gap-2'>
                      <Calendar className='h-4 w-4' />
                      <span>
                        Self evaluation starts:{' '}
                        {format(new Date(coursePhaseConfig.selfEvaluationStart), 'MMM d, yyyy')}
                      </span>
                    </div>
                  )}
                  {!peerEvaluationStarted && coursePhaseConfig?.peerEvaluationStart && (
                    <div className='flex items-center gap-2'>
                      <Calendar className='h-4 w-4' />
                      <span>
                        Peer evaluation starts:{' '}
                        {format(new Date(coursePhaseConfig.peerEvaluationStart), 'MMM d, yyyy')}
                      </span>
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <div className='space-y-3'>
                <p className='text-gray-600 dark:text-gray-300 leading-relaxed'>
                  Please complete your assigned evaluations before the specified deadlines.
                </p>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
