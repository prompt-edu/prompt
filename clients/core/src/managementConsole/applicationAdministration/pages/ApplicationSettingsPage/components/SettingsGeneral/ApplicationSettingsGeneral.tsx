import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { format, set, parse, formatISO } from 'date-fns'
import { toZonedTime } from 'date-fns-tz'
import { AlertCircle, Loader2 } from 'lucide-react'
import type { ApplicationMetaData } from '../../../../interfaces/applicationMetaData'
import type { UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import {
  Button,
  Label,
  Switch,
  DatePicker,
  Input,
  Card,
  CardContent,
} from '@tumaet/prompt-ui-components'
import { updateCoursePhase } from '@core/network/mutations/updateCoursePhase'
import { ApplicationConfigDialogError } from './ApplicationSettingsGeneralErrorMessage'

interface ApplicationConfigDialogProps {
  initialData: ApplicationMetaData
}

const getTimeString = (date: Date) => {
  return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
}

export function ApplicationGeneralSettings({ initialData }: ApplicationConfigDialogProps) {
  const queryClient = useQueryClient()
  const { phaseId } = useParams<{ phaseId: string }>()

  // States for form fields
  const [startDate, setStartDate] = useState<Date | undefined>()
  const [endDate, setEndDate] = useState<Date | undefined>()
  const [startTime, setStartTime] = useState('00:00')
  const [endTime, setEndTime] = useState('23:59')
  const [externalStudentsAllowed, setExternalStudentsAllowed] = useState(false)
  const [universityLoginAvailable, setUniversityLoginAvailable] = useState(false)
  const [autoAccept, setAutoAccept] = useState(false)
  const [useCustomScores, setUseCustomScores] = useState(false)
  const [dateError, setDateError] = useState<string | null>(null)

  const timeZone = 'Europe/Berlin'

  // Initialize form values when initialData changes
  useEffect(() => {
    setStartDate(
      initialData.applicationStartDate ? new Date(initialData.applicationStartDate) : undefined,
    )
    setEndDate(
      initialData.applicationEndDate ? new Date(initialData.applicationEndDate) : undefined,
    )
    setStartTime(
      initialData.applicationStartDate
        ? getTimeString(new Date(initialData.applicationStartDate))
        : '00:00',
    )
    setEndTime(
      initialData.applicationEndDate
        ? getTimeString(new Date(initialData.applicationEndDate))
        : '23:59',
    )
    setAutoAccept(initialData?.autoAccept ?? false)
    setExternalStudentsAllowed(initialData?.externalStudentsAllowed ?? false)
    setUniversityLoginAvailable(initialData?.universityLoginAvailable ?? false)
    setUseCustomScores(initialData?.useCustomScores ?? false)
    setDateError(null)
  }, [initialData])

  const {
    mutate: mutatePhase,
    isError: isMutateError,
    error,
    isPending,
  } = useMutation({
    mutationFn: (coursePhase: UpdateCoursePhase) => {
      return updateCoursePhase(coursePhase)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['course_phase', phaseId] })
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // Build full Date objects by combining the date and time values
    let startDateTime: Date | null = null
    let endDateTime: Date | null = null

    if (startDate) {
      const parsedStartTime = parse(startTime, 'HH:mm', new Date())
      startDateTime = set(startDate, {
        hours: parsedStartTime.getHours(),
        minutes: parsedStartTime.getMinutes(),
      })
    }
    if (endDate) {
      const parsedEndTime = parse(endTime, 'HH:mm', new Date())
      endDateTime = set(endDate, {
        hours: parsedEndTime.getHours(),
        minutes: parsedEndTime.getMinutes(),
      })
    }

    // Validate that the start date/time comes before the end date/time
    if (startDateTime && endDateTime && startDateTime.getTime() >= endDateTime.getTime()) {
      setDateError('Start date and time must be before end date and time.')
      return
    }
    // Clear any previous error
    setDateError(null)

    const updatedPhase: UpdateCoursePhase = {
      id: phaseId ?? '',
      restrictedData: {
        applicationStartDate: startDateTime
          ? formatISO(toZonedTime(startDateTime, timeZone))
          : undefined,
        applicationEndDate: endDateTime ? formatISO(toZonedTime(endDateTime, timeZone)) : undefined,
        externalStudentsAllowed,
        universityLoginAvailable,
        autoAccept,
        useCustomScores,
      },
    }

    mutatePhase(updatedPhase)
  }

  return (
    <Card className='w-full'>
      <CardContent>
        {isPending ? (
          <div className='flex items-center gap-2'>
            <Loader2 className='h-4 w-4 animate-spin' />
            <span>Saving application config...</span>
          </div>
        ) : isMutateError ? (
          <div className='space-y-4'>
            <ApplicationConfigDialogError error={error} />
          </div>
        ) : (
          <>
            <div className='mb-2 mt-5'>
              <h3 className='text-lg font-semibold'>General Settings</h3>
              <p className='text-sm text-muted-foreground'>
                Note: All times are in German time (Europe/Berlin).
              </p>
            </div>

            {/* Display validation error if present */}
            {dateError && (
              <div
                className={`flex items-center p-4 mb-4 text-sm text-red-800 border border-red-300 rounded-lg
                    bg-red-50 dark:bg-gray-800 dark:text-red-400 dark:border-red-800`}
                role='alert'
              >
                <AlertCircle className='shrink-0 inline w-4 h-4 mr-3' />
                <span className='sr-only'>Error</span>
                <div>
                  <span className='font-medium'>Validation error:</span> {dateError}
                </div>
              </div>
            )}

            <form onSubmit={handleSubmit}>
              <div className='grid gap-4 py-4'>
                {/* Start Date/Time */}
                <div className='grid grid-cols-4 items-center gap-4'>
                  <Label htmlFor='startDate' className='text-right'>
                    Start Date
                  </Label>
                  <div className='col-span-3 flex items-center gap-2'>
                    <DatePicker
                      date={startDate}
                      onSelect={(date) =>
                        setStartDate(date ? new Date(format(date, 'yyyy-MM-dd')) : undefined)
                      }
                    />
                    <Input
                      id='startTime'
                      type='time'
                      value={startTime}
                      onChange={(e) => setStartTime(e.target.value)}
                      className='w-24'
                    />
                  </div>
                </div>

                {/* End Date/Time */}
                <div className='grid grid-cols-4 items-center gap-4'>
                  <Label htmlFor='endDate' className='text-right'>
                    End Date
                  </Label>
                  <div className='col-span-3 flex items-center gap-2'>
                    <DatePicker
                      date={endDate}
                      onSelect={(date) =>
                        setEndDate(date ? new Date(format(date, 'yyyy-MM-dd')) : undefined)
                      }
                    />
                    <Input
                      id='endTime'
                      type='time'
                      value={endTime}
                      onChange={(e) => setEndTime(e.target.value)}
                      className='w-24'
                    />
                  </div>
                </div>

                {/* University Login Available */}
                <div className='grid grid-cols-4 items-center gap-4'>
                  <Label htmlFor='universityLoginAvailable' className='text-right'>
                    Enforce Student Login
                  </Label>
                  <div className='col-span-3 flex items-center gap-4'>
                    <Switch
                      id='universityLoginAvailable'
                      checked={universityLoginAvailable}
                      onCheckedChange={setUniversityLoginAvailable}
                    />
                    <p className='text-sm text-muted-foreground'>
                      This option is highly recommended. It requires a Keycloak login for students
                      and provides matriculation number and university login data.
                    </p>
                  </div>
                </div>

                {/* External Students Allowed */}
                <div className='grid grid-cols-4 items-center gap-4'>
                  <Label htmlFor='externalStudentsAllowed' className='text-right'>
                    Allow External Students
                  </Label>
                  <div className='col-span-3 flex items-center gap-4'>
                    <Switch
                      id='externalStudentsAllowed'
                      checked={externalStudentsAllowed}
                      onCheckedChange={setExternalStudentsAllowed}
                    />
                    <p className='text-sm text-muted-foreground'>
                      This option is to allow external students to apply without login and
                      matriculation number.
                    </p>
                  </div>
                </div>

                {/* Auto Accept */}
                <div className='grid grid-cols-4 items-center gap-4'>
                  <Label htmlFor='autoAccept' className='text-right'>
                    Auto Accept
                  </Label>
                  <div className='col-span-3 flex items-center gap-4'>
                    <Switch id='autoAccept' checked={autoAccept} onCheckedChange={setAutoAccept} />
                    <p className='text-sm text-muted-foreground'>
                      This option will automatically accept all applications without any manual
                      review.
                    </p>
                  </div>
                </div>
              </div>

              <div className='flex justify-end'>
                <Button type='submit'>Save changes</Button>
              </div>
            </form>
          </>
        )}
      </CardContent>
    </Card>
  )
}
