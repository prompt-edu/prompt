import { Badge } from '@tumaet/prompt-ui-components'
import { format } from 'date-fns'
import { Clock } from 'lucide-react'

import { AssessmentType } from '../../../interfaces/assessmentType'

interface DeadlineBadgeProps {
  deadline: Date | string
  type?: AssessmentType
}

export const DeadlineBadge = ({ deadline, type = AssessmentType.SELF }: DeadlineBadgeProps) => {
  const deadlineDate = typeof deadline === 'string' ? new Date(deadline) : deadline
  const currentDate = new Date()
  const isDeadlinePassed = deadlineDate < currentDate

  let badgeClassName: string
  let deadlineText: string

  if (isDeadlinePassed) {
    badgeClassName =
      'bg-red-100 text-red-800 border-red-200 dark:border-red-700 flex items-center gap-1 hover:bg-red-100'
    deadlineText = `Deadline passed: ${format(deadlineDate, 'dd.MM.yyyy')}`
  } else {
    switch (type) {
      case AssessmentType.SELF:
        badgeClassName =
          'bg-blue-100 text-blue-800 border-blue-200 dark:border-blue-700 flex items-center gap-1 hover:bg-blue-100'
        break
      case AssessmentType.PEER:
        badgeClassName =
          'bg-green-200 text-green-800 border-green-300 hover:bg-green-100 flex items-center gap-1'
        break
      case AssessmentType.TUTOR:
        badgeClassName =
          'bg-purple-200 text-purple-800 border-purple-300 hover:bg-purple-100 flex items-center gap-1'
        break
      default:
        badgeClassName =
          'bg-gray-200 text-gray-800 border-gray-300 hover:bg-gray-100 flex items-center gap-1'
        break
    }
    deadlineText = `Deadline: ${format(deadlineDate, 'dd.MM.yyyy')}`
  }

  return (
    <Badge className={badgeClassName}>
      <Clock className='h-3 w-3' />
      {deadlineText}
    </Badge>
  )
}
