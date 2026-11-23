import type { Org } from '@core/interfaces/org'
import { courseOrgID } from '@core/managementConsole/shared/utils/orgRoles'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import { CourseTable } from '../../shared/components/CourseTable/CourseTable'

interface OrgCoursesSettingsProps {
  org: Org
}

// The course list holds every course of the org for PROMPT admins and the org's Admins
export const OrgCoursesSettings = ({ org }: OrgCoursesSettingsProps) => {
  const { courses } = useCourseStore()
  const orgCourses = useMemo(
    () => courses.filter((course) => courseOrgID(course) === org.id),
    [courses, org.id],
  )

  return (
    <Card>
      <CardHeader>
        <CardTitle>Courses</CardTitle>
        <CardDescription>
          The org's courses and templates. Its Admins manage all of them; nested orgs' courses are
          not included.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {orgCourses.length === 0 ? (
          <p className='text-muted-foreground'>The org has no courses yet.</p>
        ) : (
          <CourseTable courses={orgCourses} />
        )}
      </CardContent>
    </Card>
  )
}
