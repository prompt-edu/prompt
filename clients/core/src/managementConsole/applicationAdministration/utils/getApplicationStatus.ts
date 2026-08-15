import { isAfter, isBefore, isWithinInterval } from 'date-fns'
import type { ApplicationMetaData } from '../interfaces/applicationMetaData'
import { ApplicationStatus } from '../interfaces/applicationStatus'

export const getApplicationStatus = (
  applicationMetaData: ApplicationMetaData | null,
  applicationPhaseIsConfigured: boolean,
) => {
  if (!applicationMetaData) return ApplicationStatus.Unknown
  if (applicationMetaData.applicationMode === 'import') return ApplicationStatus.Import
  if (!applicationPhaseIsConfigured) return ApplicationStatus.NotConfigured
  const now = new Date()
  if (isBefore(now, applicationMetaData.applicationStartDate!)) return ApplicationStatus.NotYetLive
  if (isAfter(now, applicationMetaData.applicationEndDate!)) return ApplicationStatus.Passed
  if (
    isWithinInterval(now, {
      start: applicationMetaData.applicationStartDate!,
      end: applicationMetaData.applicationEndDate!,
    })
  )
    return ApplicationStatus.Live
  return ApplicationStatus.Unknown
}
