import * as React from 'react'
import { differenceInDays, differenceInHours, differenceInMinutes, format } from 'date-fns'
import { Card, CardContent, CardHeader, CardTitle } from '@tumaet/prompt-ui-components'
import { ApplicationStatusBadge } from '../../../components/ApplicationStatusBadge'
import { getApplicationStatus } from '../../../utils/getApplicationStatus'
import { ApplicationMetaData } from '../../../interfaces/applicationMetaData'
import { ApplicationStatus } from '../../../interfaces/applicationStatus'

interface ApplicationStatusCardProps {
  applicationMetaData: ApplicationMetaData | null
  applicationPhaseIsConfigured: boolean
}

function getTimeUntil(date: Date, action: string) {
  const now = new Date()
  const days = differenceInDays(date, now)
  if (days >= 1) return { value: days, description: `${days === 1 ? 'day' : 'days'} ${action}` }
  const hours = differenceInHours(date, now)
  if (hours >= 1)
    return { value: hours, description: `${hours === 1 ? 'hour' : 'hours'} ${action}` }
  const minutes = Math.max(differenceInMinutes(date, now), 0)
  return { value: minutes, description: `${minutes === 1 ? 'minute' : 'minutes'} ${action}` }
}

export function ApplicationStatusCard({
  applicationMetaData,
  applicationPhaseIsConfigured,
}: ApplicationStatusCardProps) {
  const status = getApplicationStatus(applicationMetaData, applicationPhaseIsConfigured)
  const startDate = applicationMetaData?.applicationStartDate
  const endDate = applicationMetaData?.applicationEndDate

  const formatDate = (date: Date | null) => {
    return date ? format(date, 'MMM d, yyyy') : 'Not set'
  }

  const getDisplayContent = React.useMemo(() => {
    switch (status) {
      case ApplicationStatus.NotYetLive:
        return startDate
          ? getTimeUntil(startDate, 'until start')
          : { value: null, description: 'Not set' }
      case ApplicationStatus.Live:
        return endDate
          ? getTimeUntil(endDate, 'remaining')
          : { value: null, description: 'Ongoing' }
      case ApplicationStatus.Passed:
        return { value: null, description: 'CLOSED' }
      default:
        return { value: null, description: 'Unknown' }
    }
  }, [status, startDate, endDate])

  return (
    <Card className='flex flex-col h-full'>
      <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
        <CardTitle className='text-base font-medium'>Application Phase</CardTitle>
        <ApplicationStatusBadge
          applicationPhaseIsConfigured={applicationPhaseIsConfigured}
          applicationStatus={status}
        />
      </CardHeader>
      <CardContent className='flex-1 flex flex-col justify-between'>
        <div>
          {getDisplayContent.value !== null ? (
            <>
              <div className='text-6xl font-bold'>{getDisplayContent.value}</div>
              <div className='text-xl mt-1'>{getDisplayContent.description}</div>
            </>
          ) : (
            <div className='text-2xl font-bold'>{getDisplayContent.description}</div>
          )}
        </div>
        <div className='text-xs text-muted-foreground mt-4'>
          <p>{applicationPhaseIsConfigured ? 'Configured' : 'Not Configured'}</p>
          <p>Start: {formatDate(applicationMetaData?.applicationStartDate ?? null)}</p>
          <p>End: {formatDate(applicationMetaData?.applicationEndDate ?? null)}</p>
        </div>
      </CardContent>
    </Card>
  )
}
