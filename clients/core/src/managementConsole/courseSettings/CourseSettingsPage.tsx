import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { MailingConfigPage } from '../mailingConfig/MailingConfigPage'
import { CourseGeneralSettings } from './components/CourseGeneralSettings'
import CourseDangerZone from './components/CourseDangerZone'

export const CourseSettingsPage = () => {
  return (
    <div className='space-y-6'>
      <ManagementPageHeader>Settings</ManagementPageHeader>

      <CourseGeneralSettings />
      <MailingConfigPage />
      <CourseDangerZone />
    </div>
  )
}
