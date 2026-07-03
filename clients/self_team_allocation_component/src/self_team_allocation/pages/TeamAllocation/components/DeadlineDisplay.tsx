import { Alert, AlertDescription, AlertTitle } from '@tumaet/prompt-ui-components'
import { AlertCircle, Clock } from 'lucide-react'
import type { Timeframe } from '../../../interfaces/timeframe'

interface DeadlineDisplayProps {
  timeframe: Timeframe
}

export function DeadlineDisplay({ timeframe }: DeadlineDisplayProps) {
  if (!timeframe.timeframeSet) {
    return null
  }

  const now = new Date()
  const hasNotStarted = timeframe.startTime > now
  const hasEnded = timeframe.endTime < now
  const isActive = !hasNotStarted && !hasEnded

  // Format dates in a user-friendly way
  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('default', {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(date)
  }

  return (
    <div className='space-y-4'>
      <div className='flex items-center gap-2 text-sm text-muted-foreground'>
        <Clock className='h-4 w-4' />
        <span>
          {isActive ? 'Allocation active until: ' : 'Allocation period: '}
          <span className='font-medium'>
            {formatDate(timeframe.startTime)} - {formatDate(timeframe.endTime)}
          </span>
        </span>
      </div>

      {hasNotStarted && (
        <Alert>
          <AlertCircle className='h-4 w-4' />
          <AlertTitle>Allocation not yet started</AlertTitle>
          <AlertDescription>
            Team allocation will begin on {formatDate(timeframe.startTime)}. You cannot join or
            create teams until then.
          </AlertDescription>
        </Alert>
      )}

      {hasEnded && (
        <Alert variant='destructive'>
          <AlertCircle className='h-4 w-4' />
          <AlertTitle>Allocation period ended</AlertTitle>
          <AlertDescription>
            The team allocation period ended on {formatDate(timeframe.endTime)}. You can no longer
            join or create teams.
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}
