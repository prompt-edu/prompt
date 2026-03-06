import { PassStatus } from '@tumaet/prompt-shared-state'
import { ApplicationRow } from './applicationRow'
import { ColumnDef } from '@tanstack/react-table'
import { getApplicationStatusBadge } from './getApplicationStatusBadge'

export function getApplicationColumns(
  additionalScores?: { key: string; name: string }[],
): ColumnDef<ApplicationRow>[] {
  return [
    {
      id: 'firstName',
      header: 'First Name',
      cell: ({ row }) => row.original.student.firstName,
      accessorFn: (row) => row.student.firstName,
    },
    {
      id: 'lastName',
      header: 'Last Name',
      cell: ({ row }) => row.original.student.lastName,
      accessorFn: (row) => row.student.lastName,
    },
    { accessorKey: 'email', header: 'Email' },
    { accessorKey: 'studyProgram', header: 'Study Program' },
    { accessorKey: 'studyDegree', header: 'Study Degree' },
    { accessorKey: 'gender', header: 'Gender' },
    {
      accessorKey: 'passStatus',
      header: 'Status',
      cell: (info) => getApplicationStatusBadge(info.getValue() as PassStatus),
    },
    {
      accessorKey: 'score',
      header: 'Score',
      cell: (info) => info.getValue() ?? '-',
    },
    ...(additionalScores ?? []).map((s) => ({
      accessorKey: s.key,
      header: s.name,
    })),
  ]
}
