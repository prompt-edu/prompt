import { useCallback, useMemo, useState } from 'react'
import { Plus } from 'lucide-react'
import { Button, PromptTable } from '@tumaet/prompt-ui-components'
import { NoteTag, CreateNoteTag, UpdateNoteTag } from '../shared/interfaces/InstructorNote'
import {
  useNoteTags,
  useCreateNoteTag,
  useUpdateNoteTag,
  useDeleteNoteTag,
} from '@core/network/hooks/useInstructorNoteTags'
import { NoteTagFormDialog } from './components/NoteTagFormDialog'
import { noteTagTableColumns } from './components/noteTagTableColumns'
import { getNoteTagTableActions } from './components/noteTagTableActions'

export const StudentNoteTagsPage = () => {
  const { data: tags = [], isLoading, isError } = useNoteTags()
  const createTag = useCreateNoteTag()
  const updateTag = useUpdateNoteTag()
  const deleteTag = useDeleteNoteTag()

  const [createOpen, setCreateOpen] = useState(false)
  const [editTag, setEditTag] = useState<NoteTag | null>(null)

  const handleCreate = (values: CreateNoteTag) => {
    createTag.mutate(values, { onSuccess: () => setCreateOpen(false) })
  }

  const handleUpdate = (values: UpdateNoteTag) => {
    if (!editTag) return
    updateTag.mutate({ tagId: editTag.id, tag: values }, { onSuccess: () => setEditTag(null) })
  }

  const handleDelete = useCallback(
    (tags_: NoteTag[]) => {
      tags_.forEach((tag) => deleteTag.mutate(tag.id))
    },
    [deleteTag],
  )

  const actions = useMemo(
    () => getNoteTagTableActions({ onEdit: setEditTag, onDelete: handleDelete }),
    [handleDelete],
  )

  if (isLoading) return <p className='text-muted-foreground text-sm'>Loading tags…</p>
  if (isError) return <p className='text-destructive text-sm'>Failed to load tags.</p>

  return (
    <div className='flex flex-col gap-6 w-full'>
      <div className='flex items-center justify-between'>
        <h1 className='text-3xl font-bold tracking-tight'>Note Tags</h1>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className='h-4 w-4 mr-2' />
          Create Tag
        </Button>
      </div>

      <PromptTable
        data={tags}
        columns={noteTagTableColumns}
        actions={actions}
        onRowClick={setEditTag}
      />

      <NoteTagFormDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSubmit={handleCreate}
        isPending={createTag.isPending}
        title='Create Tag'
      />

      {editTag && (
        <NoteTagFormDialog
          open={!!editTag}
          onClose={() => setEditTag(null)}
          initialValues={{ name: editTag.name, color: editTag.color }}
          onSubmit={handleUpdate}
          isPending={updateTag.isPending}
          title='Edit Tag'
        />
      )}
    </div>
  )
}
