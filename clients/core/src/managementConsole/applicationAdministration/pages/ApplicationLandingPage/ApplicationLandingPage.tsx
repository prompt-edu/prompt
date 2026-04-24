import { MissingConfig } from '@tumaet/prompt-ui-components'
import { getIsApplicationConfigured } from '../../utils/getApplicationIsConfigured'
import { useState } from 'react'
import { ApplicationMetaData } from '../../interfaces/applicationMetaData'
import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useParseApplicationMetaData } from '../../hooks/useParseApplicationMetaData'
import { useApplicationStore } from '../../zustand/useApplicationStore'
import { useLocation } from 'react-router-dom'
import { useMissingConfigs } from './hooks/useMissingConfig'
import { useHideMailingWarning } from './hooks/useHideMailingWarning'
import { ApplicationStatusCard } from './diagrams/ApplicationStatusCard'
import { AssessmentDiagram } from './diagrams/AssessmentDiagram'
import { ApplicationGenderDiagram } from './diagrams/ApplicationGenderDiagram'
import { ApplicationStudyBackgroundDiagram } from './diagrams/ApplicationStudyBackgroundDiagram'
import { ApplicationStudySemesterDiagram } from './diagrams/ApplicationStudySemesterDiagram'

export const ApplicationLandingPage = () => {
  const [applicationMetaData, setApplicationMetaData] = useState<ApplicationMetaData | null>(null)
  const { pathname } = useLocation()
  const { coursePhase, participations } = useApplicationStore()
  const { hideMailingWarning } = useHideMailingWarning()
  const missingConfigs = useMissingConfigs(
    applicationMetaData,
    coursePhase,
    pathname,
    hideMailingWarning,
  )

  useParseApplicationMetaData(coursePhase, setApplicationMetaData)

  const isApplicationConfigured = getIsApplicationConfigured(applicationMetaData)

  return (
    <div>
      <ManagementPageHeader>Application Administration</ManagementPageHeader>
      <MissingConfig elements={missingConfigs} />
      <div className='grid gap-6 md:grid-cols-2 lg:grid-cols-3 mb-6'>
        <ApplicationStatusCard
          applicationMetaData={applicationMetaData}
          applicationPhaseIsConfigured={isApplicationConfigured}
        />
        <AssessmentDiagram applications={participations} />
        <ApplicationGenderDiagram applications={participations} />
      </div>
      <div className='grid gap-6 md:grid-cols-1 lg:grid-cols-2 mb-6'>
        <ApplicationStudyBackgroundDiagram applications={participations} />
        <ApplicationStudySemesterDiagram applications={participations} />
      </div>
    </div>
  )
}
