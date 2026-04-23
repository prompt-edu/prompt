import { Card, CardContent } from '@tumaet/prompt-ui-components'

import { ApplicationTimeline } from './ApplicationSettingsOverviewTimeline'
import { ApplicationSettingsOverviewHeader } from './ApplicationSettingsOverviewHeader'
import { ExternalStudentsStatusBadge } from './ApplicationSettingsOverviewExternalStudentsBadge'
import { ApplicationSettingsOverviewApplicationLink } from './ApplicationSettingsOverviewApplicationLink'
import { ApplicationMetaData } from '../../../../interfaces/applicationMetaData'

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
