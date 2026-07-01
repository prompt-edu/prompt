import React, { useState } from 'react'
import { useParseApplicationMetaData } from '../../hooks/useParseApplicationMetaData'
import type { ApplicationMetaData } from '../../interfaces/applicationMetaData'
import { getIsApplicationConfigured } from '../../utils/getApplicationIsConfigured'
import { getApplicationStatus } from '../../utils/getApplicationStatus'
import { useApplicationStore } from '../../zustand/useApplicationStore'
import ApplicationOverview from './components/Overview/ApplicationSettingsOverview'
import { ApplicationSettingsCustomScores } from './components/SettingsCustomScores/ApplicationSettingsCustomScores'
import { ApplicationGeneralSettings } from './components/SettingsGeneral/ApplicationSettingsGeneral'

export const ApplicationConfiguration = () => {
  const [applicationMetaData, setApplicationMetaData] = useState<ApplicationMetaData | null>(null)
  const { coursePhase } = useApplicationStore()

  useParseApplicationMetaData(coursePhase, setApplicationMetaData)

  const applicationPhaseIsConfigured = getIsApplicationConfigured(applicationMetaData)

  const applicationStatus = getApplicationStatus(applicationMetaData, applicationPhaseIsConfigured)

  return (
    <div className='container space-y-8'>
      <h1 className='text-4xl font-bold mb-6'>Application Settings</h1>
      <ApplicationOverview
        applicationMetaData={applicationMetaData}
        applicationPhaseIsConfigured={applicationPhaseIsConfigured}
        applicationStatus={applicationStatus}
      />

      {applicationMetaData && <ApplicationGeneralSettings initialData={applicationMetaData} />}
      {applicationMetaData && <ApplicationSettingsCustomScores initialData={applicationMetaData} />}
    </div>
  )
}
