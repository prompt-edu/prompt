import { Card, CardContent } from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../../interfaces/assessmentType'

import { useCoursePhaseConfigStore } from '../../../zustand/useCoursePhaseConfigStore'

import { AssessmentStatusBadge, DeadlineBadge } from '../../components/badges'

interface TutorEvaluationStatusCardProps {
  completedEvaluations: number
  totalEvaluations: number
  isCompleted: boolean
}

export const TutorEvaluationStatusCard = ({
  completedEvaluations,
  totalEvaluations,
  isCompleted,
}: TutorEvaluationStatusCardProps) => {
  const { coursePhaseConfig } = useCoursePhaseConfigStore()

  return (
    <Card className='border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xs'>
      <CardContent className='p-6'>
        <div className='flex items-center justify-between mb-4'>
          <h3 className='text-lg font-medium text-gray-900 dark:text-gray-100'>Tutor Evaluation</h3>
          {isCompleted ? (
            <AssessmentStatusBadge remainingAssessments={0} isFinalized={true} />
          ) : (
            <DeadlineBadge
              deadline={coursePhaseConfig?.tutorEvaluationDeadline ?? ''}
              type={AssessmentType.TUTOR}
            />
          )}
        </div>
        <div className='text-2xl font-bold text-gray-900 dark:text-gray-100 mb-1'>
          {completedEvaluations}/{totalEvaluations}
        </div>
        <p className='text-sm text-gray-500 dark:text-gray-400'>Tutors evaluated</p>
      </CardContent>
    </Card>
  )
}
