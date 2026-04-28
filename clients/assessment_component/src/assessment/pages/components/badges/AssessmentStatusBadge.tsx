import { CheckCircle, Clock, CircleCheck } from 'lucide-react'
import { cn, Badge } from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../../interfaces/assessmentType'

interface AssessmentStatusBadgeProps {
  className?: string
  remainingAssessments: number
  assessmentType?: AssessmentType
  isFinalized?: boolean
}

export function AssessmentStatusBadge({
  className,
  remainingAssessments,
  assessmentType = AssessmentType.ASSESSMENT,
  isFinalized,
}: AssessmentStatusBadgeProps) {
  const isCompleted = remainingAssessments === 0
  const isInProgress = remainingAssessments > 0
  const isCompletedButNotFinalized = isCompleted && !isFinalized

  const badgeStyles = cn(
    'items-center gap-1',
    isCompleted &&
      isFinalized &&
      cn(
        'bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-200',
        'hover:bg-green-100 hover:text-green-800 hover:dark:bg-green-800 hover:dark:text-green-200',
      ),
    isCompletedButNotFinalized &&
      cn(
        'bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-200',
        'hover:bg-blue-100 hover:text-blue-800 hover:dark:bg-blue-800 hover:dark:text-blue-200',
      ),
    isInProgress &&
      cn(
        'border-orange-200 bg-orange-100 text-orange-800 dark:border-orange-800 dark:bg-orange-950/40 dark:text-orange-200',
        'hover:bg-orange-200 hover:text-orange-900 dark:hover:bg-orange-950/55 dark:hover:text-orange-100',
      ),
    className,
  )

  return (
    <Badge className={badgeStyles} style={{ whiteSpace: 'nowrap' }}>
      {isCompleted && isFinalized && (
        <>
          <CheckCircle className='h-3.5 w-3.5' />
          <span>Completed</span>
        </>
      )}

      {isCompletedButNotFinalized && (
        <>
          <CircleCheck className='h-3.5 w-3.5' />
          <span>Ready to finalize</span>
        </>
      )}

      {isInProgress && (
        <>
          <Clock className='h-3.5 w-3.5' />
          <span>
            {remainingAssessments}{' '}
            {assessmentType === AssessmentType.ASSESSMENT
              ? remainingAssessments === 1
                ? 'assessment'
                : 'assessments'
              : remainingAssessments === 1
                ? 'question'
                : 'questions'}{' '}
            left
          </span>
        </>
      )}
    </Badge>
  )
}
