import { ManagementPageHeader, MissingConfig } from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import { useLocation } from 'react-router-dom'
import { useParseApplicationMetaData } from '../../hooks/useParseApplicationMetaData'
import type { ApplicationMetaData } from '../../interfaces/applicationMetaData'
import { getIsApplicationConfigured } from '../../utils/getApplicationIsConfigured'
import { useApplicationStore } from '../../zustand/useApplicationStore'
import { ApplicationGenderDiagram } from './diagrams/ApplicationGenderDiagram'
import { ApplicationStatusCard } from './diagrams/ApplicationStatusCard'
import { ApplicationStudyBackgroundDiagram } from './diagrams/ApplicationStudyBackgroundDiagram'
import { ApplicationStudySemesterDiagram } from './diagrams/ApplicationStudySemesterDiagram'
import { AssessmentDiagram } from './diagrams/AssessmentDiagram'
import { useHideMailingWarning } from './hooks/useHideMailingWarning'
import { useMissingConfigs } from './hooks/useMissingConfig'

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
