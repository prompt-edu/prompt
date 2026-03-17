import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { ColumnDef } from '@tanstack/react-table'
import { PromptTable } from '@tumaet/prompt-ui-components'
import { Course } from '@tumaet/prompt-shared-state'
import { ArrowRight, Settings } from 'lucide-react'
import { RowAction, TableFilter } from '@tumaet/prompt-ui-components'
import { CourseTableColumns } from './CourseTableColumns'
import { CourseTableActions } from './CourseTableActions'
import { CourseTableFilters } from './CourseTableFilters'

interface CourseTableProps {
  courses: Course[]
}

export const CourseTable = ({ courses }: CourseTableProps) => {
  const navigate = useNavigate()

  const columns: ColumnDef<Course>[] = useMemo(() => CourseTableColumns, [])

  const courseTableActions: RowAction<Course>[] = [
    ...CourseTableActions,
    {
      label: 'Open course',
      icon: <ArrowRight />,
      hide: (rows) => rows.length !== 1,
      onAction: ([course]) => {
        navigate(`/management/course/${course.id}`)
      },
    },
    {
      label: 'Open settings',
      icon: <Settings />,
      hide: (rows) => rows.length !== 1,
      onAction: ([course]) => {
        navigate(`/management/course/${course.id}/settings`)
      },
    },
  ]

  const courseTableFilters: TableFilter[] = CourseTableFilters

  const onRowClick = (course: Course) => navigate(`/management/course/${course.id}`)

  return (
    <PromptTable
      data={courses}
      columns={columns}
      onRowClick={onRowClick}
      actions={courseTableActions}
      filters={courseTableFilters}
    />
  )
}
