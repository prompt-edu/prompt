import { Badge, Card, CardContent } from '@tumaet/prompt-ui-components'
import dayjs from 'dayjs'
import { AlertCircle, CheckCircle, Clock } from 'lucide-react'
import { useEffect, useState } from 'react'

interface SurveyStatusBarProps {
  deadline: Date
  status: 'Not submitted' | 'Submitted' | 'Modified'
}

export const SurveyStatusBar = ({ deadline, status }: SurveyStatusBarProps) => {
  const [timeRemaining, setTimeRemaining] = useState<{
    days: number
    hours: number
    minutes: number
    seconds: number
    totalSeconds: number
    isPast: boolean
  }>({
    days: 0,
    hours: 0,
    minutes: 0,
    seconds: 0,
    totalSeconds: 0,
    isPast: false,
  })

  useEffect(() => {
    const calculateTimeRemaining = () => {
      const now = dayjs()
      const deadlineTime = dayjs(deadline)
      const diff = deadlineTime.diff(now, 'second')
      const isPast = diff <= 0

      if (isPast) {
        setTimeRemaining({
          days: 0,
          hours: 0,
          minutes: 0,
          seconds: 0,
          totalSeconds: 0,
          isPast: true,
        })
        return
      }

      const days = Math.floor(diff / (24 * 60 * 60))
      const hours = Math.floor((diff % (24 * 60 * 60)) / (60 * 60))
      const minutes = Math.floor((diff % (60 * 60)) / 60)
      const seconds = diff % 60

      setTimeRemaining({
        days,
        hours,
        minutes,
        seconds,
        totalSeconds: diff,
        isPast: false,
      })
    }

    calculateTimeRemaining()
    const timer = setInterval(calculateTimeRemaining, 1000)

    return () => clearInterval(timer)
  }, [deadline])

  const formatTimeUnit = (value: number, unit: string) => {
    return value > 0 ? `${value}${unit} ` : ''
  }

  const getTimeDisplay = () => {
    if (timeRemaining.isPast) {
      return 'Deadline passed'
    }

    if (timeRemaining.days > 0) {
      return `${formatTimeUnit(timeRemaining.days, 'd')}${formatTimeUnit(timeRemaining.hours, 'h')} remaining`
    }

    if (timeRemaining.hours > 0) {
      return `${formatTimeUnit(timeRemaining.hours, 'h')}${formatTimeUnit(timeRemaining.minutes, 'm')} remaining`
    }

    return `${formatTimeUnit(timeRemaining.minutes, 'm')}${formatTimeUnit(timeRemaining.seconds, 's')} remaining`
  }

  const getStatusColor = () => {
    switch (status) {
      case 'Submitted':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'Modified':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'Not submitted':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  const getStatusIcon = () => {
    switch (status) {
      case 'Submitted':
        return <CheckCircle className='h-4 w-4 mr-1' />
      case 'Modified':
        return <AlertCircle className='h-4 w-4 mr-1' />
      case 'Not submitted':
        return <AlertCircle className='h-4 w-4 mr-1' />
      default:
        return null
    }
  }

  return (
    <Card className='mb-6'>
      <CardContent className='p-4'>
        <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
          <div className='space-y-2'>
            <div className='flex items-center text-sm text-muted-foreground'>
              <Clock className='h-4 w-4 mr-1' />
              <span>Deadline</span>
            </div>
            <div>
              <h3 className='text-lg font-semibold'>
                {dayjs(deadline).format('MMM D, YYYY [at] h:mm A')}
              </h3>
              <div className='mt-1 text-sm font-medium'>{getTimeDisplay()}</div>
            </div>
          </div>

          <div className='space-y-2'>
            <div className='text-sm text-muted-foreground'>Status</div>
            <div className='flex items-center'>
              <Badge
                variant='outline'
                className={`${getStatusColor()} flex items-center px-3 py-1`}
              >
                {getStatusIcon()}
                {status}
              </Badge>
            </div>
            <div className='text-sm text-muted-foreground mt-1'>
              {status === 'Submitted' && 'Your response has been recorded.'}
              {status === 'Modified' &&
                'You have unsaved changes. Please submit to save your response.'}
              {status === 'Not submitted' && 'Please complete and submit the survey.'}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
