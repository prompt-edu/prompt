import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@tumaet/prompt-ui-components'
import { Course } from '@tumaet/prompt-shared-state'
import { CalendarDays, GraduationCap, Clock, Calendar, ChevronRight } from 'lucide-react'
import { CourseTypeDetails } from '@tumaet/prompt-shared-state'
import DynamicIcon from '@/components/DynamicIcon'
import { formatDate } from '@core/utils/formatDate'
import { CourseArchiveButton } from './CourseArchiveButton'
import { useNavigate } from 'react-router-dom'
import { CourseSettingsButton } from './CourseSettingsButton'

type CourseMetaItemProps = {
  icon: React.ReactNode
  label: string
  value: React.ReactNode
}

const CourseMetaItem = ({ icon, label, value }: CourseMetaItemProps) => (
  <div className='flex items-center gap-3'>
    <div className='bg-gray-100 p-2 rounded-full'>{icon}</div>
    <div>
      <p className='text-sm font-medium text-gray-500'>{label}</p>
      <p className='text-base font-semibold'>{value}</p>
    </div>
  </div>
)

type CourseCardProps = {
  course: Course
}

export const CourseCard = ({ course }: CourseCardProps) => {
  const bgColor = course.studentReadableData?.['bg-color'] || 'bg-gray-50'
  const navigate = useNavigate()

  return (
    <Card className='overflow-hidden border border-gray-200 h-full flex flex-col'>
      <CardHeader className={`rounded-t-lg ${bgColor} py-6 px-6 border-b`}>
        <div className='flex items-center justify-between gap-4'>
          <div className='flex items-center gap-4 min-w-0'>
            <div className='size-6 shrink-0'>
              <DynamicIcon
                name={course.studentReadableData?.['icon'] || 'graduation-cap'}
                color='black'
              />
            </div>

            <CardTitle
              className='text-2xl font-bold text-gray-900 leading-tight truncate'
              title={course.name}
            >
              {course.name}
            </CardTitle>
          </div>

          <div className='flex gap-2'>
            <CourseArchiveButton courseID={course.id} archived={course.archived} />
            <CourseSettingsButton courseID={course.id} />
          </div>
        </div>
      </CardHeader>

      <CardContent className='p-6 flex-grow'>
        <div className='space-y-5'>
          <CourseMetaItem
            icon={<CalendarDays className='w-5 h-5 text-gray-700' />}
            label='Semester'
            value={course.semesterTag}
          />

          <CourseMetaItem
            icon={<Calendar className='w-5 h-5 text-gray-700' />}
            label='Duration'
            value={
              course.startDate && course.endDate
                ? `${formatDate(course.startDate.toString())} - ${formatDate(course.endDate.toString())}`
                : 'N/A'
            }
          />

          <CourseMetaItem
            icon={<GraduationCap className='w-5 h-5 text-gray-700' />}
            label='Course Type'
            value={CourseTypeDetails[course.courseType].name}
          />

          <CourseMetaItem
            icon={<Clock className='w-5 h-5 text-gray-700' />}
            label='ECTS'
            value={course.ects}
          />
        </div>
      </CardContent>

      <CardFooter className='px-6 py-4 border-t flex justify-between'>
        <div className='text-xs text-gray-500'>
          {course.archivedOn && <>Archived on {formatDate(course.archivedOn)}</>}
        </div>

        <button
          onClick={() => navigate(`/management/course/${course.id}`)}
          className='text-sm font-medium text-primary flex items-center hover:underline focus-visible:ring-2 focus-visible:ring-offset-2 rounded'
        >
          Go to course <ChevronRight className='ml-1 h-4 w-4' />
        </button>
      </CardFooter>
    </Card>
  )
}
