import { useState } from 'react'
import { InstructorNote as InstructorNoteType } from '../../interfaces/InstructorNote'
import {
  useDeleteInstructorNote,
  useCreateInstructorNote,
} from '@core/network/hooks/useInstructorNotes'
import { NoteWrapper } from './InstructorNoteWrapper'
import { NoteComposer } from './InstructorNoteComposer'
import { NoteVersionHistoryItem } from './InstructorNoteVersionHistoryItem'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Button,
} from '@tumaet/prompt-ui-components'

interface NoteProps {
  note: InstructorNoteType
  studentId: string
}

export function InstructorNote({ note, studentId }: NoteProps) {
  const deleteNote = useDeleteInstructorNote(studentId)
  const createNote = useCreateInstructorNote(studentId)

  const [isEditing, setIsEditing] = useState(false)
  const [showVersions, setShowVersions] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const isDeleted = note.dateDeleted != null
  const latestVersion = note.versions[note.versions.length - 1]
  const olderVersions = latestVersion ? note.versions.slice(0, -1).reverse() : []

  const handleDelete = () => {
    setShowDeleteDialog(true)
  }

  const handleConfirmDelete = () => {
    deleteNote.mutate(note.id)
    setShowDeleteDialog(false)
  }

  return (
    <>
      <NoteWrapper
        note={note}
        isEditing={isEditing}
        onEdit={() => setIsEditing(true)}
        onDelete={handleDelete}
        isDeleting={deleteNote.isPending}
        showVersions={showVersions}
        onToggleVersions={() => setShowVersions((p) => !p)}
      >
        {isDeleted ? (
          <p className='text-sm text-muted-foreground italic'>This note was deleted</p>
        ) : isEditing ? (
          <NoteComposer
            initialContent={latestVersion?.content ?? ''}
            initialTags={note.tags}
            onSubmit={async (content, tagIds) => {
              await createNote.mutateAsync({ content, new: false, forNote: note.id, tags: tagIds })
              setIsEditing(false)
            }}
            onCancel={() => setIsEditing(false)}
            isPending={createNote.isPending}
            autoFocus
          />
        ) : (
          <p className='text-sm whitespace-pre-wrap'>{latestVersion?.content ?? ''}</p>
        )}

        {!isEditing && showVersions && olderVersions.length > 0 && (
          <div className='mt-2 space-y-2 border-l-2 border-gray-200 pl-3'>
            {olderVersions.map((noteVersion) => (
              <NoteVersionHistoryItem
                key={noteVersion.id}
                content={noteVersion.content}
                dateCreated={noteVersion.dateCreated}
              />
            ))}
          </div>
        )}
      </NoteWrapper>

      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className='sm:max-w-[400px]'>
          <DialogHeader>
            <DialogTitle>Delete Note</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this note? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant='outline' onClick={() => setShowDeleteDialog(false)}>
              Cancel
            </Button>
            <Button
              variant='destructive'
              onClick={handleConfirmDelete}
              disabled={deleteNote.isPending}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
