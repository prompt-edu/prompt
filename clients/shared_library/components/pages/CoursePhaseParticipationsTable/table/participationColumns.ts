import { ColumnDef } from '@tanstack/react-table'
import { ExtraParticipantColumn, ParticipantRow } from './participationRow'
import { getStatusBadge } from '@/utils/getStatusBadge'
import { PassStatus } from '@tumaet/prompt-shared-state'

export function getParticipantColumns(
  extraColumns: ExtraParticipantColumn<any>[],
): ColumnDef<ParticipantRow, any>[] {
  return [
    {
      accessorKey: 'firstName',
      header: 'First name',
    },
    {
      accessorKey: 'lastName',
      header: 'Last name',
    },
    {
      accessorKey: 'matriculationNumber',
      header: 'Matriculation #',
    },
    {
      accessorKey: 'universityLogin',
      header: 'Login',
    },
    {
      accessorKey: 'passStatus',
      header: 'Status',
      cell: (info) => getStatusBadge(info.getValue() as PassStatus),
    },

    ...extraColumns.map(
      (col): ColumnDef<ParticipantRow, any> => ({
        id: col.id,
        header: col.header,
        accessorFn: col.accessorFn!,
        cell: col.cell,
        enableSorting: col.enableSorting,
        sortingFn: col.sortingFn,
        enableColumnFilter: col.enableColumnFilter,
        filterFn: col.filterFn,
      }),
    ),
  ]
}
