import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Input,
  Label,
} from '@tumaet/prompt-ui-components'
import dayjs from 'dayjs'
import timezone from 'dayjs/plugin/timezone'
import utc from 'dayjs/plugin/utc'
import { AlertCircle, CalendarIcon, Loader2 } from 'lucide-react'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import type { SurveyTimeframe } from '../../../interfaces/timeframe'
import { updateSurveyTimeframe as updateSurveyTimeframeFn } from '../../../network/mutations/updateSurveyTimeframe'

dayjs.extend(utc)
dayjs.extend(timezone)

interface SurveyTimeframeSettingsProps {
  surveyTimeframe: SurveyTimeframe
}

export const SurveyTimeframeSettings = ({ surveyTimeframe }: SurveyTimeframeSettingsProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  // We'll assume the user's local time zone as reported by the browser.
  // If you'd prefer a fixed/project-wide time zone (e.g. 'America/Los_Angeles'),
  // you can hard-code that instead.
  const userTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone

  /**
   * Convert a UTC Date to a 'YYYY-MM-DDTHH:mm' string in the user’s local time zone,
   * so that <input type='datetime-local' /> will display the correct local time.
   */
  const formatDateTimeForInput = (date: Date) => {
    return dayjs(date).tz(userTimeZone).format('YYYY-MM-DDTHH:mm')
  }

  /**
   * When the user picks a date+time in <input type='datetime-local' />,
   * it comes through as 'YYYY-MM-DDTHH:mm' (no time zone).
   * We interpret it in the user’s local time zone, then convert to a UTC Date object.
   */
  const parseLocalDateTimeToUtc = (localDateTimeString: string) => {
    return dayjs.tz(localDateTimeString, userTimeZone).utc().toDate()
  }

  const [startDateTime, setStartDateTime] = useState<string>(
    surveyTimeframe.timeframeSet ? formatDateTimeForInput(surveyTimeframe.surveyStart) : '',
  )
  const [endDateTime, setEndDateTime] = useState<string>(
    surveyTimeframe.timeframeSet ? formatDateTimeForInput(surveyTimeframe.surveyDeadline) : '',
  )

  const [error, setError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: ({ start, end }: { start: Date; end: Date }) =>
      updateSurveyTimeframeFn(phaseId ?? '', start, end),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team_allocation_survey_timeframe', phaseId] })
      setError(null)
    },
    onError: () => {
      setError('Failed to update survey timeframe. Please try again.')
    },
  })

  const handleUpdate = (e: React.FormEvent) => {
    e.preventDefault()

    // Convert local input to UTC
    const start = parseLocalDateTimeToUtc(startDateTime)
    const end = parseLocalDateTimeToUtc(endDateTime)

    if (start > end) {
      setError('Start date cannot be after end date.')
      return
    }

    setError(null)
    mutation.mutate({ start, end })
  }

  return (
    <Card className='w-full'>
      <CardHeader>
        <CardTitle className='text-xl flex items-center gap-2'>
          <CalendarIcon className='h-5 w-5 text-primary' />
          Survey Timeframe
        </CardTitle>
        <CardDescription>
          Set the start/end date & time (with timezone handling) for the survey.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className='space-y-6'>
          <form id='timeframe-form' onSubmit={handleUpdate} className='space-y-6'>
            <div className='grid sm:grid-cols-2 gap-6'>
              <div className='space-y-2'>
                <Label htmlFor='start-date-time' className='font-medium'>
                  Start Date & Time
                </Label>
                <div className='relative'>
                  <Input
                    id='start-date-time'
                    type='datetime-local'
                    value={startDateTime}
                    onChange={(e) => setStartDateTime(e.target.value)}
                    className='w-full'
                    required
                  />
                </div>
              </div>
              <div className='space-y-2'>
                <Label htmlFor='end-date-time' className='font-medium'>
                  End Date & Time
                </Label>
                <div className='relative'>
                  <Input
                    id='end-date-time'
                    type='datetime-local'
                    value={endDateTime}
                    onChange={(e) => setEndDateTime(e.target.value)}
                    className='w-full'
                    required
                  />
                </div>
              </div>
            </div>

            {error && (
              <Alert variant='destructive' className='mt-4'>
                <AlertCircle className='h-4 w-4' />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
          </form>

          <div className='flex justify-end'>
            <Button
              type='submit'
              form='timeframe-form'
              disabled={
                mutation.isPending ||
                !startDateTime ||
                !endDateTime ||
                // Compare the local-formatted versions of the original dates
                (startDateTime === formatDateTimeForInput(surveyTimeframe.surveyStart) &&
                  endDateTime === formatDateTimeForInput(surveyTimeframe.surveyDeadline))
              }
              className='min-w-[180px]'
            >
              {mutation.isPending ? (
                <>
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  Updating...
                </>
              ) : (
                'Update Timeframe'
              )}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
