import {
  DropdownMenuCheckboxItem,
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
} from '@tumaet/prompt-ui-components'
import { Tag } from 'lucide-react'
import { InstructorNoteTag } from '../InstructorNoteTag'
import { NoteTag } from '../../../interfaces/InstructorNote'
import { useNoteTags } from '@core/network/hooks/useInstructorNoteTags'

interface InstructorNoteComposerTagPickerProps {
  selectedTags: NoteTag[]
  onToggle: (tag: NoteTag) => void
}

const iconButtonClass =
  'shrink-0 p-1 rounded text-muted-foreground hover:bg-muted hover:text-foreground'

export function InstructorNoteComposerTagPicker({
  selectedTags,
  onToggle,
}: InstructorNoteComposerTagPickerProps) {
  const { data: availableTags = [] } = useNoteTags()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className={iconButtonClass}>
          <Tag className='w-4 h-4' />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className='w-auto p-1' align='start'>
        {availableTags.length === 0 ? (
          <p className='text-xs text-muted-foreground px-2 py-1'>No tags available</p>
        ) : (
          availableTags.map((tag) => {
            const isSelected = selectedTags.some((t) => t.id === tag.id)
            return (
              <DropdownMenuCheckboxItem
                key={tag.id}
                onClick={() => onToggle(tag)}
                checked={isSelected}
              >
                <InstructorNoteTag tag={tag} />
              </DropdownMenuCheckboxItem>
            )
          })
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
