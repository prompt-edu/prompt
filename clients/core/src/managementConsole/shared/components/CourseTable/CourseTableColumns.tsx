import { Course, CourseTypeDetails } from '@tumaet/prompt-shared-state'
import DynamicIcon from '@/components/DynamicIcon'
import { formatDate } from '@core/utils/formatDate'
import { ColumnDef } from '@tanstack/react-table'

export const CourseTableColumns: ColumnDef<Course>[] = [
  {
    id: 'icon',
    header: 'Icon',
    enableSorting: false,
    cell: ({ row }) => {
      const iconName = row.original.studentReadableData?.['icon'] || 'graduation-cap'
      const bgColor = row.original.studentReadableData?.['bg-color'] || 'bg-gray-50'
      return (
        <div
          className={`inline-flex items-center justify-center rounded-md ${bgColor} text-black`}
          style={{ width: 24, height: 24 }}
        >
          <DynamicIcon name={iconName} className='h-4 w-4' />
        </div>
      )
    },
    size: 80,
  },
  {
    accessorKey: 'name',
    header: 'Name',
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: 'semesterTag',
    header: 'Semester',
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: 'courseType',
    header: 'Course Type',
    cell: (info) => CourseTypeDetails[info.getValue() as Course['courseType']].name,
    filterFn: 'arrIncludes',
  },
  {
    accessorKey: 'startDate',
    header: 'Start Date',
    cell: (info) => formatDate(info.getValue() as string | Date),
  },
  {
    accessorKey: 'endDate',
    header: 'End Date',
    cell: (info) => formatDate(info.getValue() as string | Date),
  },
  {
    accessorKey: 'ects',
    header: 'ECTS',
    cell: (info) => info.getValue(),
  },
  {
    id: 'status',
    header: 'Status',
    enableSorting: true,
    accessorFn: (course) => {
      if (course.template) return 'Template'
      if (course.archived) return 'Archived'
      return 'Active'
    },
    sortingFn: (rowA, rowB) => {
      const a = rowA.original.archivedOn
      const b = rowB.original.archivedOn
      if (a && b) {
        return new Date(b).getTime() - new Date(a).getTime()
      }
      if (a && !b) return -1
      if (!a && b) return 1
      return 0
    },
    cell: (info) => {
      const status = info.getValue<string>()
      if (status === 'Archived') {
        const archivedOn = info.row.original.archivedOn
        return archivedOn ? `Archived on ${formatDate(archivedOn)}` : 'Archived'
      }
      return status
    },
  },
]
