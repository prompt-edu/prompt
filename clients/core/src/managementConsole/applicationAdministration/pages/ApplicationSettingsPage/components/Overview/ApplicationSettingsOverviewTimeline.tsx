import * as React from 'react'
import { format } from 'date-fns'
import { Progress } from '@tumaet/prompt-ui-components'
import { formatTimeRemaining } from '../../../../utils/formatTimeRemaining'

interface ApplicationTimelineProps {
  startDate: Date | undefined
  endDate: Date | undefined
}

export const ApplicationTimeline: React.FC<ApplicationTimelineProps> = ({
  startDate,
  endDate,
}: ApplicationTimelineProps) => {
  if (!startDate || !endDate) return <div>Unknown Dates</div>

  const today = new Date()
  const totalMs = endDate.getTime() - startDate.getTime()
  const passedMs = today.getTime() - startDate.getTime()
  const progress = Math.min(Math.max((passedMs / totalMs) * 100, 0), 100)

  return (
    <div className='space-y-2'>
      <div className='flex justify-between text-sm text-gray-500'>
        <span>{format(startDate, 'MMM d, yyyy')}</span>
        <span>{format(endDate, 'MMM d, yyyy')}</span>
      </div>
      <Progress value={progress} className='w-full' />
      <div className='text-center text-sm text-gray-500'>
        {progress <= 0 ? (
          <span>Application period starts in {formatTimeRemaining(startDate, today)}</span>
        ) : progress >= 100 ? (
          <span>Application period has ended</span>
        ) : (
          <span>{formatTimeRemaining(endDate, today)} remaining</span>
        )}
      </div>
    </div>
  )
}
