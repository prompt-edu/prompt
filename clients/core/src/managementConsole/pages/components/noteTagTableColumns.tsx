import { ColumnDef } from '@tanstack/react-table'
import { InstructorNoteTagColor } from '../../shared/components/InstructorNote/InstructorNoteTag'
import { NoteTag } from '../../shared/interfaces/InstructorNote'

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
