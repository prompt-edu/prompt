import { ReactNode, useState } from 'react'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import { InstructorNote } from '../../interfaces/InstructorNote'
import { formatNoteDate } from '@core/utils/formatDate'
import { ProfilePicture } from '@/components/StudentProfilePicture'
import { NoteActionButtons } from './InstructorNoteActionButtons'
import { InstructorNoteTags } from './InstructorNoteTag'

interface NoteWrapperProps {
  note: InstructorNote
  isEditing: boolean
  onEdit: () => void
  onDelete: () => void
  isDeleting: boolean
  showVersions?: boolean
  onToggleVersions?: () => void
  children: ReactNode
}

export function NoteWrapper({
  note,
  isEditing,
  onEdit,
  onDelete,
  isDeleting,
  showVersions,
  onToggleVersions,
  children,
}: NoteWrapperProps) {
  const { user } = useAuthStore()
  const [isHovered, setIsHovered] = useState(false)

  const isOwner = user?.email === note.authorEmail
  const isDeleted = note.dateDeleted != null
  const names = note.authorName.split(' ')
  const latestVersion = note.versions[note.versions.length - 1]

  return (
    <div
      className={`relative flex gap-3 transition-all rounded-2xl ${showVersions ? 'bg-gray-50 p-2' : 'p-0'}`}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <ProfilePicture
        email={note.authorEmail}
        firstName={names[0]}
        lastName={names[names.length - 1]}
        size='md'
      />

      <div className='flex-1 min-w-0' style={{ lineHeight: '11px' }}>
        <div className='text-sm flex items-center'>
          <span className='font-semibold'>{note.authorName}</span>
          {!isEditing && !isDeleted && note.versions.length > 1 && onToggleVersions && (
            <>
              <span className='mx-1'>·</span>
              <button className='hover:text-blue-500' onClick={onToggleVersions}>
                {showVersions ? 'hide edits' : 'edited'}
              </button>
            </>
          )}
        </div>

        <div className='flex-col gap-1'>
          {!isEditing && <InstructorNoteTags tags={note.tags} />}
          {children}
        </div>
      </div>

      {!showVersions && (
        <div
          className={`absolute right-0 top-0 flex items-center gap-1 px-1 h-6
          bg-background [box-shadow:-12px_0_12px_2px_hsl(var(--background))] [clip-path:inset(0_0_0_-50px)] transform -translate-y-[2px]
          transition-opacity ${isHovered && !isEditing ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}
        >
          {latestVersion && (
            <span className='text-muted-foreground font-normal text-xs'>
              {formatNoteDate(note.dateCreated)}
            </span>
          )}
          {isOwner && !isDeleted && (
            <NoteActionButtons
              isVisible={true}
              onEdit={onEdit}
              onDelete={onDelete}
              isDeleting={isDeleting}
            />
          )}
        </div>
      )}
    </div>
  )
}
