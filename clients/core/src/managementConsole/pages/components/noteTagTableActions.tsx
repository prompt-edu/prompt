import { Pencil, Trash2 } from 'lucide-react'
import { RowAction } from '@tumaet/prompt-ui-components'
import { NoteTag } from '../../shared/interfaces/InstructorNote'

interface NoteTagTableActionsProps {
  onEdit: (tag: NoteTag) => void
  onDelete: (tags: NoteTag[]) => void
}

export function getNoteTagTableActions({
  onEdit,
  onDelete,
}: NoteTagTableActionsProps): RowAction<NoteTag>[] {
  return [
    {
      label: 'Edit',
      icon: <Pencil />,
      onAction: (rows) => onEdit(rows[0]),
      hide: (rows) => rows.length > 1,
    },
    {
      label: 'Delete',
      icon: <Trash2 />,
      onAction: (rows) => onDelete(rows),
      confirm: {
        title: 'Delete Tag',
        description: (count) =>
          `Are you sure you want to delete ${count} tag${count > 1 ? 's' : ''}? This will remove them from all notes that use them.`,
        confirmLabel: 'Delete',
        variant: 'destructive',
      },
    },
  ]
}
