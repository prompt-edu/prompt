import { format } from 'date-fns'
import { getChallengeStatusBadge } from './utils/getChallengeStatusBadge'
import { ChallengeStatus } from './interfaces/challengeStatus'
import { Badge, Checkbox } from '@tumaet/prompt-ui-components'
import { SortableHeader } from '@/components/table/SortableHeader'
import { getStatusBadge } from '@/utils/getStatusBadge'

export const columns = [
  {
    id: 'select',
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label='Select all'
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onClick={(event) => {
          event.stopPropagation()
          row.toggleSelected()
        }}
        aria-label='Select row'
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: 'participation.student.lastName',
    header: ({ column }) => <SortableHeader column={column} title='Name' />,
    cell: ({ row }) =>
      `${row.original.participation?.student?.firstName || ''} ${row.original.participation?.student?.lastName || ''}`,
    sortingFn: (rowA, rowB) => {
      const valueA = `${
        rowA.original.participation?.student?.lastName?.toLowerCase() || ''
      } ${rowA.original.participation?.student?.firstName?.toLowerCase() || ''}`
      const valueB = `${
        rowB.original.participation?.student?.lastName?.toLowerCase() || ''
      } ${rowB.original.participation?.student?.firstName?.toLowerCase() || ''}`
      return valueA.localeCompare(valueB)
    },
  },
  {
    accessorKey: 'participation.student.email',
    header: ({ column }) => <SortableHeader column={column} title='Mail' />,
    cell: ({ row }) => row.original.participation?.student?.email || '',
  },
  {
    accessorKey: 'passStatus',
    header: ({ column }) => <SortableHeader column={column} title='Pass Status' />,
    cell: ({ row }) => getStatusBadge(row.original.participation?.passStatus),
    filterFn: (row, columnId, filterValue) => {
      return filterValue.includes(row.original.participation?.passStatus)
    },
  },
  {
    accessorKey: 'challengeStatus',
    header: ({ column }) => <SortableHeader column={column} title='Challenge Status' />,
    cell: ({ row }) => {
      if (row.original.profile === undefined) {
        return getChallengeStatusBadge(ChallengeStatus.NOT_STARTED)
      } else if (row.original.profile.hasPassed) {
        return getChallengeStatusBadge(ChallengeStatus.PASSED)
      }
      return getChallengeStatusBadge(ChallengeStatus.NOT_COMPLETED)
    },
    sortingFn: (rowA, rowB) => {
      if (rowA.original.profile === rowB.original.profile) {
        return 0
      } else if (rowA.original.profile?.hasPassed) {
        return 1
      }
      return -1
    },
    filterFn: (row, columnId, filterValue) => {
      return filterValue.includes(
        row.original.profile?.hasPassed === undefined
          ? ChallengeStatus.NOT_STARTED
          : row.original.profile?.hasPassed
            ? ChallengeStatus.PASSED
            : ChallengeStatus.NOT_COMPLETED,
      )
    },
  },
  {
    accessorKey: 'passingPosition',
    header: ({ column }) => <SortableHeader column={column} title='Position' />,
    cell: ({ row }) => row.original.profile?.passingPosition ?? '-',
    sortingFn: (rowA, rowB) => {
      const valueA = rowA.original.profile?.passingPosition ?? -1
      const valueB = rowB.original.profile?.passingPosition ?? -1
      return valueA - valueB
    },
  },
  {
    accessorKey: 'passedAt',
    header: ({ column }) => <SortableHeader column={column} title='Passed At' />,
    cell: ({ row }) =>
      row.original.profile?.passedAt && row.original.profile?.passedAt instanceof Date ? (
        <Badge variant='outline'>
          {format(row.original.profile?.passedAt, 'dd.MM.yyyy HH:mm')}
        </Badge>
      ) : (
        '-'
      ),
    sortingFn: (rowA, rowB) => {
      const valueA =
        rowA.original.profile?.passedAt instanceof Date
          ? rowA.original.profile.passedAt.toISOString()
          : ''
      const valueB =
        rowB.original.profile?.passedAt instanceof Date
          ? rowB.original.profile.passedAt.toISOString()
          : ''
      return valueA.localeCompare(valueB)
    },
  },
  {
    accessorKey: 'attempts',
    header: ({ column }) => <SortableHeader column={column} title='Attempts' />,
    cell: ({ row }) => row.original.profile?.attempts ?? '-',
    sortingFn: (rowA, rowB) => {
      const valueA = rowA.original.profile?.attempts ?? -1
      const valueB = rowB.original.profile?.attempts ?? -1
      return valueA - valueB
    },
  },
]
