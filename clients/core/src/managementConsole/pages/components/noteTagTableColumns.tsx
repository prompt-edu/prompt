import type { PromptTableColumnDef } from '@tumaet/prompt-ui-components'
import { InstructorNoteTagColor } from '../../shared/components/InstructorNote/InstructorNoteTag'
import type { NoteTag } from '../../shared/interfaces/InstructorNote'

export const noteTagTableColumns: PromptTableColumnDef<NoteTag>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
  },
  {
    accessorKey: 'color',
    header: 'Color',
    cell: ({ row }) => <InstructorNoteTagColor color={row.original.color} />,
  },
]
