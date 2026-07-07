import { StudentAvatar } from '@tumaet/prompt-ui-components'

interface RequesterCellProps {
  user_id: string
  student_id: string | null
  student_first_name: string | null
  student_last_name: string | null
  student_email: string | null
}

export function RequesterCell({
  student_id,
  student_first_name,
  student_last_name,
  student_email,
}: RequesterCellProps) {
  if (student_id) {
    return (
      <StudentAvatar
        student={{
          id: student_id,
          firstName: student_first_name ?? '',
          lastName: student_last_name ?? '',
          email: student_email ?? '',
        }}
      />
    )
  }

  return <span className='text-sm text-muted-foreground'>Non-student user / removed</span>
}

export function requesterAccessor(row: {
  student_first_name: string | null
  student_last_name: string | null
  student_email: string | null
  user_id: string
}) {
  const studentLabel = [row.student_first_name, row.student_last_name, row.student_email]
    .filter(Boolean)
    .join(' ')
  return studentLabel || row.user_id
}
