import { CourseAvatar } from '@core/managementConsole/layout/Sidebar/CourseSwitchSidebar/components/CourseAvatar'
import { StudentCourseParticipation } from '@core/network/queries/getStudentsWithCourses'

export const StudentCoursePreview = ({
  studentCourseParticipation,
}: {
  studentCourseParticipation: StudentCourseParticipation
}) => {
  return (
    <div className={`flex items-center text-black text-sm mr-1`}>
      <div style={{ zoom: 0.85 }}>
        <CourseAvatar
          bgColor={studentCourseParticipation.studentReadableData['bg-color']}
          iconName={studentCourseParticipation.studentReadableData['icon']}
          className='size-8 mr-1'
        />
      </div>
      {studentCourseParticipation.courseName}
    </div>
  )
}
