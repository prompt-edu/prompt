import {
  getStudentsWithCourses,
  StudentWithCourses,
} from '@core/network/queries/getStudentsWithCourses'
import { ColumnDef } from '@tanstack/react-table'
import { PromptTable, RowAction, TableFilter } from '@tumaet/prompt-ui-components'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { studentTableColumns } from './studentTableColumns'
import { getStudentTableFilters } from './studentTableFilters'
import { getStudentTableActions } from './studentTableActions'
import { useStudentStore } from '../../store/student.store'

export const StudentTable = () => {
  const [studentsWithCourses, setStudentsWithCourses] = useState<Array<StudentWithCourses>>([])

  const { upsertStudents } = useStudentStore()

  const navigate = useNavigate()
  const openStudent = useCallback(
    (student: StudentWithCourses) => navigate(`/management/students/${student.id}`),
    [navigate],
  )

  useEffect(() => {
    const fetchStudents = async () => {
      const s = await getStudentsWithCourses()
      setStudentsWithCourses(s)
      upsertStudents(s)
    }
    fetchStudents()
  }, [upsertStudents])

  const columns: ColumnDef<StudentWithCourses>[] = useMemo(() => studentTableColumns, [])

  const filters: TableFilter[] = useMemo(
    () => getStudentTableFilters(studentsWithCourses),
    [studentsWithCourses],
  )

  const actions: RowAction<StudentWithCourses>[] = useMemo(
    () => getStudentTableActions({ openStudent }),
    [openStudent],
  )
  return (
    <div className='flex flex-col gap-3 w-full'>
      <PromptTable
        data={studentsWithCourses}
        columns={columns}
        filters={filters}
        actions={actions}
        onRowClick={openStudent}
      />
    </div>
  )
}
