import { useParams } from 'react-router-dom'
import { useStudent } from '@core/network/hooks/useStudent'
import { Loader2 } from 'lucide-react'
import { StudentProfile } from '@/components/StudentProfile'
import { InstructorNotes } from '../shared/components/InstructorNote/InstructorNotes'
import { StudentDetailContentLayout } from '../shared/components/StudentDetail/StudentDetailContentLayout'
import { CourseEnrollments } from '../shared/components/StudentDetail/CourseEnrollmentList'
import { EmptyPage } from '../shared/components/EmptyPage'

export const StudentDetailPage = () => {
  const { studentId } = useParams<{ studentId: string }>()

  const student = useStudent(studentId)

  if (!studentId) {
    return <EmptyPage message='Student not found' />
  }

  return (
    <div className='flex flex-col w-full justify-between gap-2'>
      <div className='flex flex-col gap-y-2 text-sm'>
        {student.isLoading && <Loader2 className='h-4 w-4 animate-spin' />}
        {student.isError && <p className='text-destructive'>Failed to load student data</p>}
        {student.isSuccess && <StudentProfile student={student.data} />}
      </div>

      <StudentDetailContentLayout
        courseEnrollment={<CourseEnrollments studentId={studentId} />}
        instructorNotes={<InstructorNotes studentId={studentId} />}
      />
    </div>
  )
}
