import { StudentWithCourses } from '@core/network/queries/getStudentsWithCourses'
import { RowAction } from '@tumaet/prompt-ui-components'
import { ArrowRight } from 'lucide-react'

interface passedFunctions {
  openStudent: (student: StudentWithCourses) => void
}

function hideWhenMulti(selectedRows: StudentWithCourses[]) {
  return selectedRows.length > 1
}

export function getStudentTableActions({
  openStudent,
}: passedFunctions): RowAction<StudentWithCourses>[] {
  return [
    {
      label: 'Open Student',
      icon: <ArrowRight />,
      onAction: (selectedRows: StudentWithCourses[]) => {
        openStudent(selectedRows[0])
      },
      hide: hideWhenMulti,
    },
  ]
}
