import { useCourseStore } from '@tumaet/prompt-shared-state'
import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'
import { MailingConfigPage } from '../mailingConfig/MailingConfigPage'
import { courseOrgID } from '../shared/utils/orgRoles'
import CourseDangerZone from './components/CourseDangerZone'
import { CourseGeneralSettings } from './components/CourseGeneralSettings'
import { CourseOrgSettings } from './components/CourseOrgSettings'

export const CourseSettingsPage = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const { courses } = useCourseStore()
  const course = courses.find((c) => c.id === courseId)

  return (
    <div className='space-y-6'>
      <ManagementPageHeader>Settings</ManagementPageHeader>

      <CourseGeneralSettings />
      {course && (
        // Keyed by the saved org, so a saved change resets the selection
        <CourseOrgSettings key={`${course.id}-${courseOrgID(course) ?? 'none'}`} course={course} />
      )}
      <MailingConfigPage />
      <CourseDangerZone />
    </div>
  )
}
