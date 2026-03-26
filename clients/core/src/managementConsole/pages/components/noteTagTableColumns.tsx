import { ColumnDef } from '@tanstack/react-table'
import { NoteTag } from '../../shared/interfaces/InstructorNote'
import { InstructorNoteTagColor } from '../../shared/components/InstructorNote/InstructorNoteTag'

export const noteTagTableColumns: ColumnDef<NoteTag>[] = [
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
