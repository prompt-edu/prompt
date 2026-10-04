import { AIKeySettings } from '@managementConsole/ai/components/AIKeySettings'
import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { MailingConfigPage } from '../mailingConfig/MailingConfigPage'
import CourseDangerZone from './components/CourseDangerZone'
import { CourseGeneralSettings } from './components/CourseGeneralSettings'

export const CourseSettingsPage = () => {
  return (
    <div className='space-y-6'>
      <ManagementPageHeader>Settings</ManagementPageHeader>

      <CourseGeneralSettings />
      <MailingConfigPage />
      <AIKeySettings />
      <CourseDangerZone />
    </div>
  )
}
