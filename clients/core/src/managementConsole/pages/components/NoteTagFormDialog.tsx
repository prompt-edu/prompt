import { useState } from 'react'
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { NoteTagColor } from '../../shared/interfaces/InstructorNote'
import {
  InstructorNoteTag,
  InstructorNoteTagColor,
} from '../../shared/components/InstructorNote/InstructorNoteTag'

const NOTE_TAG_COLORS: NoteTagColor[] = ['blue', 'green', 'red', 'yellow', 'orange', 'pink']

interface NoteTagFormDialogProps {
  open: boolean
  onClose: () => void
  initialValues?: { name: string; color: NoteTagColor }
  onSubmit: (values: { name: string; color: NoteTagColor }) => void
  isPending: boolean
  title: string
}

export const NoteTagFormDialog = ({
  open,
  onClose,
  initialValues,
  onSubmit,
  isPending,
  title,
}: NoteTagFormDialogProps) => {
  const [name, setName] = useState(initialValues?.name ?? '')
  const [color, setColor] = useState<NoteTagColor>(initialValues?.color ?? 'blue')

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) onClose()
  }

  const handleSubmit = () => {
    if (name.trim()) {
      onSubmit({ name: name.trim(), color })
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <div className='flex justify-center py-2'>
          <div style={{ zoom: 1.15 }}>
            <InstructorNoteTag tag={{ id: '', name: name.trim() || 'Tag preview', color }} />
          </div>
        </div>
        <div className='flex flex-col gap-4 py-2'>
          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='tag-name'>Name</Label>
            <Input
              id='tag-name'
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder='Tag name'
            />
          </div>
          <div className='flex flex-col gap-1.5'>
            <Label>Color</Label>
            <Select value={color} onValueChange={(v) => setColor(v as NoteTagColor)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {NOTE_TAG_COLORS.map((c) => (
                  <SelectItem key={c} value={c}>
                    <InstructorNoteTagColor color={c} />
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant='outline' onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!name.trim() || isPending}>
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
