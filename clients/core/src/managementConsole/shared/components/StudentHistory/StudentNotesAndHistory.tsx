import { InstructorNotes } from '@core/managementConsole/shared/components/InstructorNote/InstructorNotes'
import { ShowForRole } from '@core/managementConsole/shared/components/ShowForRole'
import { CourseEnrollments } from '@core/managementConsole/shared/components/StudentDetail/CourseEnrollmentList'
import { Role } from '@tumaet/prompt-shared-state'
import { Card } from '@tumaet/prompt-ui-components'

export function StudentNotesAndHistory({ studentId }: { studentId: string }) {
  return (
    <>
      <ShowForRole roles={[Role.PROMPT_ADMIN, Role.PROMPT_LECTURER]}>
        <Card className='p-3'>
          <InstructorNotes studentId={studentId} />
        </Card>
      </ShowForRole>
      <Card className='p-3'>
        <CourseEnrollments studentId={studentId} />
      </Card>
    </>
  )
}
