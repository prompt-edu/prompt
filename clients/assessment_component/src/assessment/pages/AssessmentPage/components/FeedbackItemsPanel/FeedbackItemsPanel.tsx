import { ErrorPage } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'

import { useCoursePhaseConfigStore } from '../../../../zustand/useCoursePhaseConfigStore'
import { useStudentAssessmentStore } from '../../../../zustand/useStudentAssessmentStore'

import { FeedbackItemDisplayPanel } from '../../../components/FeedbackItemDisplayPanel/FeedbackItemDisplayPanel'

import { useGetFeedbackItemsForStudent } from './hooks/useGetFeedbackItemsForStudent'

export const FeedbackItemsPanel = () => {
  const { assessmentParticipation } = useStudentAssessmentStore()
  const courseParticipationID = assessmentParticipation?.courseParticipationID || ''

  const { positiveFeedbackItems, negativeFeedbackItems, isLoading, isError, refetch } =
    useGetFeedbackItemsForStudent(courseParticipationID)
  const { coursePhaseConfig } = useCoursePhaseConfigStore()

  const studentName = assessmentParticipation?.student?.firstName || 'this student'

  const getHeading = () => {
    const selfEnabled = coursePhaseConfig?.selfEvaluationEnabled
    const peerEnabled = coursePhaseConfig?.peerEvaluationEnabled

    if (selfEnabled && peerEnabled) {
      return 'Feedback Items from Self and Peer Evaluation'
    } else if (selfEnabled) {
      return 'Feedback Items from Self Evaluation'
    } else if (peerEnabled) {
      return 'Feedback Items from Peer Evaluation'
    } else {
      return 'Feedback Items'
    }
  }

  if (isError) {
    return <ErrorPage message='Error loading feedback items' onRetry={refetch} />
  }

  if (isLoading) {
    return (
      <div className='flex justify-center items-center h-64'>
        <div role='status' aria-label='Loading feedback items'>
          <Loader2 className='h-12 w-12 animate-spin text-primary' />
          <span className='sr-only'>Loading feedback items...</span>
        </div>
      </div>
    )
  }

  if (positiveFeedbackItems.length === 0 && negativeFeedbackItems.length === 0) {
    return (
      <div className='space-y-4'>
        <h1 className='text-xl font-semibold tracking-tight'>{getHeading()}</h1>
        <div className='flex flex-col items-center justify-center p-8 text-center border rounded-lg bg-muted/30'>
          <p className='text-muted-foreground my-2'>No feedback items available for this student</p>
        </div>
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <h1 className='text-xl font-semibold tracking-tight'>{getHeading()}</h1>

      <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
        <FeedbackItemDisplayPanel
          feedbackItems={negativeFeedbackItems}
          feedbackType='negative'
          studentName={studentName}
        />
        <FeedbackItemDisplayPanel
          feedbackItems={positiveFeedbackItems}
          feedbackType='positive'
          studentName={studentName}
        />
      </div>
    </div>
  )
}
