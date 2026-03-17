import { useMemo } from 'react'
import { CalendarX, MailWarningIcon } from 'lucide-react'
import type { MissingConfigItem } from '@/components/MissingConfig'
import { getIsApplicationConfigured } from '../../../utils/getApplicationIsConfigured'
import type { ApplicationMetaData } from '../../../interfaces/applicationMetaData'

export const useMissingConfigs = (
  applicationMetaData: ApplicationMetaData | null,
  coursePhase: any,
  path: string,
  hideMailingWarning: () => void,
): MissingConfigItem[] => {
  return useMemo(() => {
    const missingConfigItems: MissingConfigItem[] = []

    if (!getIsApplicationConfigured(applicationMetaData)) {
      missingConfigItems.push({
        title: 'Application Phase Deadlines',
        icon: CalendarX,
        link: `${path}/settings`,
      })
    }

    if (
      coursePhase?.restrictedData?.mailingSettings === undefined &&
      !coursePhase?.restrictedData?.hideMailingWarning
    ) {
      missingConfigItems.push({
        title: 'Application Mailing Settings',
        description: `This application phase has no mailing settings configured.
          If you do not want to send mails, you can hide this warning.`,
        icon: MailWarningIcon,
        link: `${path}/mailing`,
        hide: hideMailingWarning,
      })
    }

    return missingConfigItems
  }, [
    applicationMetaData,
    coursePhase?.restrictedData?.hideMailingWarning,
    coursePhase?.restrictedData?.mailingSettings,
    hideMailingWarning,
    path,
  ])
}
