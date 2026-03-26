import { useAuthStore } from '@tumaet/prompt-shared-state'
import { useCreateInstructorNote } from '@core/network/hooks/useInstructorNotes'
import { ProfilePicture } from '@/components/StudentProfilePicture'
import { NoteComposer } from './InstructorNoteComposer'

interface InstructorNotesCreateFormProps {
  studentId: string
  className?: string
}

export function InstructorNotesCreateForm({
  studentId,
  className = '',
}: InstructorNotesCreateFormProps) {
  const { user } = useAuthStore()
  const createNote = useCreateInstructorNote(studentId)

  return (
    <div className={`flex gap-3 ${className}`}>
      <div>
        <ProfilePicture
          email={user?.email ?? ''}
          firstName={user?.firstName ?? ''}
          lastName={user?.lastName ?? ''}
        />
      </div>
      <div className='w-full'>
        <div className='text-sm font-medium'>
          <span className='font-semibold'>
            {user?.firstName} {user?.lastName}
          </span>
        </div>
        <NoteComposer
          onSubmit={async (content, tagIds) => {
            await createNote.mutateAsync({ content, new: true, tags: tagIds })
          }}
          isPending={createNote.isPending}
        />
      </div>
    </div>
  )
}
