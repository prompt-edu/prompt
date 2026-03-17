import {
  getPermissionString,
  Role,
  useAuthStore,
  useCourseStore,
} from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from '@tumaet/prompt-ui-components'
import { CalendarDays, GraduationCap, Clock, Calendar } from 'lucide-react'
import { CourseTypeDetails } from '@tumaet/prompt-shared-state'
import { CourseStatusTag } from '../layout/Sidebar/CourseSwitchSidebar/components/CourseStatusTag'
import type { Course } from '@tumaet/prompt-shared-state'
import { CourseSettingsButton } from '../shared/components/CourseCard/CourseSettingsButton'

export const CourseOverview = () => {
  const { courses } = useCourseStore()
  const { courseId } = useParams<{ courseId: string }>()
  const course = courses.find((c) => c.id === courseId) as Course | undefined
  const { permissions } = useAuthStore()

  const formatDate = (dateString: string): string => {
    const [year, month, date] = dateString.split('-')
    return `${date}.${month}.${year}`
  }

  const canEdit =
    permissions.includes(
      getPermissionString(Role.COURSE_LECTURER, course?.name, course?.semesterTag),
    ) || permissions.includes(Role.PROMPT_ADMIN)

  const bgColor = course?.studentReadableData?.['bg-color'] || 'bg-gray-50'

  if (!course) {
    return (
      <div className='flex items-center justify-center h-screen bg-gray-50'>
        <Card className='w-full max-w-md shadow-lg'>
          <CardHeader>
            <CardTitle className='text-center text-red-600'>Course Not Found</CardTitle>
          </CardHeader>
          <CardContent>
            <p className='text-center'>
              The requested course could not be found. Please check the course ID and try again.
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className='container mx-auto px-4 py-8'>
      <Card className='mb-8 shadow-lg'>
        <CardHeader className={`rounded-t-lg ${bgColor}`}>
          <div className='flex justify-between items-center'>
            <div>
              <CardTitle className='text-3xl font-bold text-black'>{course.name}</CardTitle>
              <CardDescription className='mt-2 text-black'>
                <CourseStatusTag course={course} />
              </CardDescription>
            </div>

            {canEdit && <CourseSettingsButton courseID={course.id} />}
          </div>
        </CardHeader>
        <CardContent className='p-6'>
          {course.shortDescription && (
            <p className='mb-4 text-muted-foreground leading-relaxed'>{course.shortDescription}</p>
          )}
          {course.longDescription && (
            <p className='mb-6 text-sm text-muted-foreground leading-relaxed border-l-2 border-border pl-3'>
              {course.longDescription}
            </p>
          )}
          <div className='grid md:grid-cols-2 gap-6'>
            <div className='space-y-4'>
              <div className='flex items-center space-x-3'>
                <CalendarDays className='w-6 h-6' />
                <div>
                  <p className='text-secondary-foreground'>Semester</p>
                  <p className='text-lg'>{course.semesterTag}</p>
                </div>
              </div>
              <div className='flex items-center space-x-3'>
                <Calendar className='w-6 h-6' />
                <div>
                  <p className='text-secondary-foreground'>Duration</p>
                  <p className='text-lg'>
                    {course.startDate && course.endDate
                      ? `${formatDate(course.startDate.toString())} - ${formatDate(course.endDate.toString())}`
                      : 'N/A'}
                  </p>
                </div>
              </div>
            </div>
            <div className='space-y-4'>
              <div className='flex items-center space-x-3'>
                <GraduationCap className='w-6 h-6' />
                <div>
                  <p className='text-secondary-foreground'>Course Type</p>
                  <p className='text-lg'>{CourseTypeDetails[course.courseType].name}</p>
                </div>
              </div>
              <div className='flex items-center space-x-3'>
                <Clock className='w-6 h-6 ' />
                <div>
                  <p className='text-secondary-foreground'>ECTS</p>
                  <p className='text-lg'>{course.ects}</p>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
