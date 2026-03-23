import { ManagementPageHeader } from '@tumaet/prompt-ui-components'

import { CoursePhaseConfigSelection } from './components/CoursePhaseConfigSelection/CoursePhaseConfigSelection'

export const SettingsPage = () => {
  return (
    <div className='space-y-6'>
      <div className='space-y-2'>
        <ManagementPageHeader>Assessment Settings</ManagementPageHeader>
        <p className='max-w-3xl text-sm leading-6 text-muted-foreground'>
          Configure the final assessment and each optional evaluation flow separately. Every card
          owns its schema, timeframe, and workflow settings, while schema authoring now lives on a
          dedicated detail page.
        </p>
      </div>

      <CoursePhaseConfigSelection />
    </div>
  )
}
