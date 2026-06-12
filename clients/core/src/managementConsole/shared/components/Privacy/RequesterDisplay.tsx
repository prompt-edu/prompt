interface RequesterDisplayProps {
  user_id: string
  student_first_name: string | null
  student_last_name: string | null
  student_email: string | null
}

export function RequesterDisplay({
  user_id,
  student_first_name,
  student_last_name,
  student_email,
}: RequesterDisplayProps) {
  const name = [student_first_name, student_last_name].filter(Boolean).join(' ')
  if (name) {
    return (
      <div className='flex flex-col text-sm'>
        <span>{name}</span>
        {student_email && <span className='text-muted-foreground text-xs'>{student_email}</span>}
      </div>
    )
  }
  return <span className='text-muted-foreground text-sm'>user {user_id.slice(0, 8)}</span>
}
