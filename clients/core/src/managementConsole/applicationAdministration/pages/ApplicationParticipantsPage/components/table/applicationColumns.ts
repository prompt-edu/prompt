import type { ExportedAnswerColumn } from '@core/managementConsole/applicationAdministration/interfaces/exportedApplicationAnswers'
import type { ColumnDef } from '@tanstack/react-table'
import type { PassStatus } from '@tumaet/prompt-shared-state'
import { type ApplicationRow, EXPORTED_ANSWER_COLUMN_PREFIX } from './applicationRow'
import { getApplicationStatusBadge } from './getApplicationStatusBadge'

export function getApplicationColumns(
  additionalScores?: { key: string; name: string }[],
  exportedColumns?: ExportedAnswerColumn[],
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
    ...(exportedColumns ?? []).map(
      (column): ColumnDef<ApplicationRow> => ({
        id: `${EXPORTED_ANSWER_COLUMN_PREFIX}${column.questionID}`,
        accessorKey: `${EXPORTED_ANSWER_COLUMN_PREFIX}${column.questionID}`,
        header: column.title,
        cell: (info) => (info.getValue() as string | undefined) ?? '-',
      }),
    ),
  ]
}
