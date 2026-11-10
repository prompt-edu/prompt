import { useState } from 'react'
import { useParseApplicationMetaData } from '../../hooks/useParseApplicationMetaData'
import type { ApplicationMetaData } from '../../interfaces/applicationMetaData'
import { getIsApplicationConfigured } from '../../utils/getApplicationIsConfigured'
import { getApplicationStatus } from '../../utils/getApplicationStatus'
import { useApplicationStore } from '../../zustand/useApplicationStore'
import { ApplicationMailingSettings } from '../Mailing/ApplicationMailingSettings'
import ApplicationOverview from './components/Overview/ApplicationSettingsOverview'
import { ApplicationSettingsCustomScores } from './components/SettingsCustomScores/ApplicationSettingsCustomScores'
import { ApplicationGeneralSettings } from './components/SettingsGeneral/ApplicationSettingsGeneral'
import { ApplicationSettingsPreview } from './components/SettingsPreview/ApplicationSettingsPreview'
import { ApplicationSettingsWelcomeText } from './components/SettingsWelcomeText/ApplicationSettingsWelcomeText'

export const ApplicationConfiguration = () => {
  const [applicationMetaData, setApplicationMetaData] = useState<ApplicationMetaData | null>(null)
  const { coursePhase } = useApplicationStore()

  useParseApplicationMetaData(coursePhase, setApplicationMetaData)

  const applicationPhaseIsConfigured = getIsApplicationConfigured(applicationMetaData)

  const applicationStatus = getApplicationStatus(applicationMetaData, applicationPhaseIsConfigured)

  return (
    <div className='container space-y-8'>
      <div className='mb-6 flex items-center justify-between gap-4'>
        <h1 className='text-4xl font-bold'>Application Settings</h1>
        <ApplicationSettingsPreview welcomeText={applicationMetaData?.welcomeText} />
      </div>
      <ApplicationOverview
        applicationMetaData={applicationMetaData}
        applicationPhaseIsConfigured={applicationPhaseIsConfigured}
        applicationStatus={applicationStatus}
      />

      {applicationMetaData && <ApplicationGeneralSettings initialData={applicationMetaData} />}
      {applicationMetaData && <ApplicationSettingsWelcomeText initialData={applicationMetaData} />}
      {applicationMetaData && <ApplicationSettingsCustomScores initialData={applicationMetaData} />}
      <ApplicationMailingSettings />
    </div>
  )
}
