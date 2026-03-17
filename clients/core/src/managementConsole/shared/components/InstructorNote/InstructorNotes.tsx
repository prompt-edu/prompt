import { useInstructorNotes } from '@core/network/hooks/useInstructorNotes'
import { InstructorNote } from './InstructorNote'
import { InstructorNotesCreateForm } from './InstructorNoteCreateForm'

interface InstructorNotesProps {
  studentId: string
}

export function InstructorNotes({ studentId }: InstructorNotesProps) {
  const notes = useInstructorNotes(studentId)

  if (notes.isError) {
    return <p>error fetching instructor notes</p>
  }

  if (notes.isPending) {
    return <></>
  }

  if (notes.isSuccess) {
    return (
      <div className='flex flex-col p-2 gap-3'>
        {notes.data.map((note) => (
          <InstructorNote key={note.id} note={note} studentId={studentId} />
        ))}
        {notes.data.length > 0 && <hr className='mt-3 mb-1' />}
        <InstructorNotesCreateForm studentId={studentId} />
      </div>
    )
  }
}
