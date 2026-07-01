import { Card, CardContent } from '@tumaet/prompt-ui-components'
import type { ApplicationMetaData } from '../../../../interfaces/applicationMetaData'
import { ApplicationSettingsOverviewApplicationLink } from './ApplicationSettingsOverviewApplicationLink'
import { ExternalStudentsStatusBadge } from './ApplicationSettingsOverviewExternalStudentsBadge'
import { ApplicationSettingsOverviewHeader } from './ApplicationSettingsOverviewHeader'
import { ApplicationTimeline } from './ApplicationSettingsOverviewTimeline'

type Props = {
  applicationMetaData: ApplicationMetaData | null
  applicationPhaseIsConfigured: boolean
  applicationStatus: string
}

const ApplicationOverview = ({
  applicationMetaData,
  applicationPhaseIsConfigured,
  applicationStatus,
}: Props) => {
  return (
    <Card className='w-full'>
      <ApplicationSettingsOverviewHeader
        applicationPhaseIsConfigured={applicationPhaseIsConfigured ?? false}
        applicationStatus={applicationStatus}
      />
      <CardContent className='space-y-6'>
        {applicationPhaseIsConfigured && (
          <>
            <ApplicationTimeline
              startDate={applicationMetaData?.applicationStartDate}
              endDate={applicationMetaData?.applicationEndDate}
            />
            <ExternalStudentsStatusBadge
              externalStudentsAllowed={applicationMetaData?.externalStudentsAllowed ?? false}
            />
            <ApplicationSettingsOverviewApplicationLink />
          </>
        )}
      </CardContent>
    </Card>
  )
}

export default ApplicationOverview
