import { useStudentEnrollments } from '@core/network/hooks/useStudentEnrollments'
import { Loader2 } from 'lucide-react'
import { CourseEnrollment } from '../../interfaces/StudentEnrollment'
import { CourseEnrollmentSummary } from './CourseEnrollmentSummary'
import { StudentCourseEnrollment } from './StudentCourseEnrollment'

interface CourseEnrollmentsProps {
  studentId: string
}

export function CourseEnrollments({ studentId }: CourseEnrollmentsProps) {
  const enrollments = useStudentEnrollments(studentId)

  return (
    <div className='flex flex-col gap-5'>
      {enrollments.isLoading && <Loader2 className='h-4 w-4 animate-spin' />}
      {enrollments.isError && <p className='text-destructive'>Failed to load enrollments</p>}
      {enrollments.isSuccess && (
        <>
          {enrollments.data?.courses.map((ce: CourseEnrollment) => (
            <div className='flex gap-4 w-full' key={ce.courseId}>
              <StudentCourseEnrollment courseEnrollment={ce} studentId={studentId} />
            </div>
          ))}
          <CourseEnrollmentSummary enrollments={enrollments.data?.courses || []} />
        </>
      )}
    </div>
  )
}
