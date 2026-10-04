import type { PromptTableColumnDef } from '@tumaet/prompt-ui-components'
import type { StudyProgramWithStudentCount } from '../utils/studyProgramStudents'

export const studyProgramTableColumns: PromptTableColumnDef<StudyProgramWithStudentCount>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
  },
  {
    accessorKey: 'shortName',
    header: 'Short Name',
    cell: ({ row }) => row.original.shortName ?? '—',
  },
  {
    accessorKey: 'studentCount',
    header: 'Students',
    cell: ({ row }) => row.original.studentCount ?? '—',
  },
]
