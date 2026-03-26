import { CourseAvatar } from '@core/managementConsole/layout/Sidebar/CourseSwitchSidebar/components/CourseAvatar'
import { LinkHeading } from './LinkHeading'
import { formatDate } from './util/formatDate'
import { CourseEnrollment } from '../../interfaces/StudentEnrollment'

interface CourseDetailProps {
  studentCourseEnrollment: CourseEnrollment
}

export function CourseDetail({ studentCourseEnrollment }: CourseDetailProps) {
  return (
    <div className='flex gap-2'>
      <CourseAvatar
        bgColor={studentCourseEnrollment.studentReadableData['bg-color']}
        iconName={studentCourseEnrollment.studentReadableData['icon']}
      />
      <div>
        <div className='flex items-baseline gap-1'>
          <LinkHeading targetURL={`/management/course/${studentCourseEnrollment.courseId}`}>
            <h3 className='font-semibold text-xl leading-tight'>{studentCourseEnrollment.name}</h3>
          </LinkHeading>
        </div>
        <p className='text-sm text-muted-foreground'></p>

        <p className='text-sm text-muted-foreground'>
          <span>{studentCourseEnrollment.semesterTag}</span> ·{' '}
          <span className='capitalize'>{studentCourseEnrollment.courseType}</span> ·{' '}
          <span>{studentCourseEnrollment.ects} ECTS</span>
        </p>

        <p className='text-sm text-muted-foreground'>
          {formatDate(studentCourseEnrollment.startDate)} -{' '}
          {formatDate(studentCourseEnrollment.endDate)}
        </p>
      </div>
    </div>
  )
}
