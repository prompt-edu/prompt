import { differenceInDays, differenceInHours, differenceInMinutes, format } from 'date-fns'
import { Progress } from '@tumaet/prompt-ui-components'

interface ApplicationTimelineProps {
  startDate: Date | undefined
  endDate: Date | undefined
}

export const ApplicationTimeline: React.FC<ApplicationTimelineProps> = ({
  startDate,
  endDate,
}: ApplicationTimelineProps) => {
  if (!startDate || !endDate) return <div>Unknown Dates</div>

  const totalDays = differenceInDays(endDate, startDate)
  const today = new Date()
  const daysPassed = differenceInDays(today, startDate)
  const progress = Math.min(Math.max((daysPassed / totalDays) * 100, 0), 100)

  return (
    <div className='space-y-2'>
      <div className='flex justify-between text-sm text-gray-500'>
        <span>{format(startDate, 'MMM d, yyyy')}</span>
        <span>{format(endDate, 'MMM d, yyyy')}</span>
      </div>
      <Progress value={progress} className='w-full' />
      <div className='text-center text-sm text-gray-500'>
        {progress <= 0 ? (
          (() => {
            const hoursUntil = differenceInHours(startDate, today)
            if (hoursUntil >= 24) return <span>Application period starts in {-daysPassed} days</span>
            const minutesUntil = differenceInMinutes(startDate, today)
            if (hoursUntil >= 1) return <span>Application period starts in {hoursUntil} hours</span>
            return <span>Application period starts in {Math.max(minutesUntil, 0)} minutes</span>
          })()
        ) : progress >= 100 ? (
          <span>Application period has ended</span>
        ) : (
          (() => {
            const daysLeft = totalDays - daysPassed
            if (daysLeft >= 1) return <span>{daysLeft} days remaining</span>
            const hoursLeft = differenceInHours(endDate, today)
            if (hoursLeft >= 1) return <span>{hoursLeft} hours remaining</span>
            const minutesLeft = differenceInMinutes(endDate, today)
            return <span>{Math.max(minutesLeft, 0)} minutes remaining</span>
          })()
        )}
      </div>
    </div>
  )
}
