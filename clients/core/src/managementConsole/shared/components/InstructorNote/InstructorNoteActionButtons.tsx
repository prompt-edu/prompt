import { Button } from '@tumaet/prompt-ui-components'
import { Trash2, Pencil } from 'lucide-react'

interface NoteActionButtonsProps {
  isVisible: boolean
  onEdit: () => void
  onDelete: () => void
  isDeleting: boolean
}

export function NoteActionButtons({
  isVisible,
  onEdit,
  onDelete,
  isDeleting,
}: NoteActionButtonsProps) {
  return (
    <div
      className={`flex gap-1 transition-opacity ${
        isVisible ? 'opacity-100' : 'opacity-0 pointer-events-none'
      }`}
    >
      <Button
        variant='ghost'
        size='icon'
        className='h-6 w-6 text-muted-foreground hover:text-blue-500'
        onClick={onEdit}
        aria-label='Edit note'
        title='Edit note'
      >
        <Pencil className='h-4 w-4' />
      </Button>
      <Button
        variant='ghost'
        size='icon'
        className='h-6 w-6 text-muted-foreground hover:text-destructive'
        onClick={onDelete}
        disabled={isDeleting}
        aria-label='Delete note'
        title='Delete note'
      >
        <Trash2 className='h-4 w-4' />
      </Button>
    </div>
  )
}
