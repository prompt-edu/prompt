import type { StudentWithCourses } from '@core/network/queries/getStudentsWithCourses'
import type { RowAction } from '@tumaet/prompt-ui-components'
import { ArrowRight, Trash2 } from 'lucide-react'

interface passedFunctions {
  openStudent: (student: StudentWithCourses) => void
  onInitiateDeletion?: (students: StudentWithCourses[]) => void
}

function hideWhenMulti(selectedRows: StudentWithCourses[]) {
  return selectedRows.length > 1
}

export function getStudentTableActions({
  openStudent,
  onInitiateDeletion,
}: passedFunctions): RowAction<StudentWithCourses>[] {
  const actions: RowAction<StudentWithCourses>[] = [
    {
      label: 'Open Student',
      icon: <ArrowRight />,
      onAction: (selectedRows: StudentWithCourses[]) => {
        openStudent(selectedRows[0])
      },
      hide: hideWhenMulti,
    },
  ]

  if (onInitiateDeletion) {
    actions.push({
      label: 'Delete Student Data',
      icon: <Trash2 />,
      onAction: (selectedRows: StudentWithCourses[]) => {
        onInitiateDeletion(selectedRows)
      },
    })
  }

  return actions
}
