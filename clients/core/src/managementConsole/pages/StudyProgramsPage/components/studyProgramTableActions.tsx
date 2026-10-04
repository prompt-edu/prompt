import type { RowAction } from '@tumaet/prompt-ui-components'
import { Pencil, Trash2 } from 'lucide-react'
import type { StudyProgramWithStudentCount } from '../utils/studyProgramStudents'

interface StudyProgramTableActionsProps {
  onEdit: (studyProgram: StudyProgramWithStudentCount) => void
  onDelete: (studyPrograms: StudyProgramWithStudentCount[]) => void
}

export function getStudyProgramTableActions({
  onEdit,
  onDelete,
}: StudyProgramTableActionsProps): RowAction<StudyProgramWithStudentCount>[] {
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
    },
  ]
}
