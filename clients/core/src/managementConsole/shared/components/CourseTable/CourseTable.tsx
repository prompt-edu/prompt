import { ColumnDef } from '@tanstack/react-table'
import { Course, Role } from '@tumaet/prompt-shared-state'
import { PromptTable, RowAction, TableFilter } from '@tumaet/prompt-ui-components'
import { ArrowRight, Settings } from 'lucide-react'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useHasRolePermission } from '../ShowForRole'
import { CourseTableActions } from './CourseTableActions'
import { CourseTableColumns } from './CourseTableColumns'
import { CourseTableFilters } from './CourseTableFilters'

interface CourseTableProps {
  courses: Course[]
}

export const CourseTable = ({ courses }: CourseTableProps) => {
  const navigate = useNavigate()
  const showActions = useHasRolePermission({ roles: [Role.PROMPT_ADMIN, Role.PROMPT_LECTURER] })

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
      actions={showActions ? courseTableActions : undefined}
      filters={courseTableFilters}
    />
  )
}
