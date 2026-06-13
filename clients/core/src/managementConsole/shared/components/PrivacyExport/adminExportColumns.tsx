import {
  type AdminExportDoc,
  type AdminPrivacyExport,
  ExportStatus,
} from '@core/network/queries/privacyStudentDataExport'
import { ColumnDef } from '@tanstack/react-table'
import { HoverInfoText } from '../Privacy/HoverInfoText'
import { StudentAvatar } from '@tumaet/prompt-ui-components'

function CountWithTooltip({ docs, label }: { docs: AdminExportDoc[]; label: string }) {
  if (docs.length === 0) return <span className='text-muted-foreground'>0 {label}</span>
  return (
    <HoverInfoText
      content={
        <ul className='text-xs space-y-0.5'>
          {docs.map((d) => (
            <li key={d.source_name}>{d.source_name}</li>
          ))}
        </ul>
      }
    >
      {docs.length} {label}
    </HoverInfoText>
  )
}

function DocSummaryCell({ docs }: { docs: AdminExportDoc[] }) {
  if (docs.length === 0) return <span className='text-muted-foreground'>no docs</span>

  const succeeded = docs.filter(
    (d) => d.status === ExportStatus.complete || d.status === ExportStatus.archived,
  )
  const noData = docs.filter((d) => d.status === ExportStatus.no_data)
  const failed = docs.filter((d) => d.status === ExportStatus.failed)

  return (
    <div className='flex flex-col gap-0.5 text-sm'>
      <span className='text-muted-foreground text-xs'>{docs.length} total</span>
      <div className='flex flex-wrap gap-x-3 gap-y-0.5'>
        <CountWithTooltip docs={succeeded} label='ok' />
        {noData.length > 0 && <CountWithTooltip docs={noData} label='no data' />}
        {failed.length > 0 && <CountWithTooltip docs={failed} label='failed' />}
      </div>
    </div>
  )
}

export const exportStatusLabel: Record<ExportStatus, string> = {
  [ExportStatus.pending]: 'Pending',
  [ExportStatus.complete]: 'Completed',
  [ExportStatus.no_data]: 'No data',
  [ExportStatus.failed]: 'Failed',
  [ExportStatus.archived]: 'Archived',
}

export const adminExportColumns: ColumnDef<AdminPrivacyExport>[] = [
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => exportStatusLabel[row.original.status],
  },
  {
    id: 'requester',
    header: 'Requester',
    accessorFn: (row) =>
      [row.student_first_name, row.student_last_name, row.student_email].filter(Boolean).join(' '),
    cell: ({ row }) => (
      <StudentAvatar
        student={{
          id: row.original.student_id ?? undefined,
          firstName: row.original.student_first_name ?? '',
          lastName: row.original.student_last_name ?? '',
          email: row.original.student_email ?? '',
        }}
      />
    ),
  },
  {
    accessorKey: 'date_created',
    header: 'Requested',
    cell: (info) => new Date(info.getValue<string>()).toLocaleString(),
  },
  {
    accessorKey: 'valid_until',
    header: 'Validity',
    cell: ({ row }) => {
      const validUntil = new Date(row.original.valid_until)
      const now = new Date()
      if (validUntil < now) {
        return (
          <div className='flex flex-col text-sm'>
            <span>Expired</span>
            <span className='text-muted-foreground text-xs'>
              on {validUntil.toLocaleDateString()}
            </span>
          </div>
        )
      }
      const daysLeft = Math.ceil((validUntil.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
      return (
        <span className='text-sm'>
          valid for {daysLeft} more {daysLeft === 1 ? 'day' : 'days'}
        </span>
      )
    },
  },
  {
    id: 'documents',
    header: 'Documents',
    cell: ({ row }) => <DocSummaryCell docs={row.original.docs} />,
  },
  {
    id: 'downloaded',
    header: 'Downloaded',
    cell: ({ row }) => {
      const docs = row.original.docs
      const downloadable = docs.filter(
        (d) => d.status === ExportStatus.complete || d.status === ExportStatus.archived,
      )
      const downloaded = downloadable.filter((d) => d.downloaded)
      if (downloadable.length === 0) return <span className='text-muted-foreground'>-</span>
      return <CountWithTooltip docs={downloaded} label={`/ ${downloadable.length} downloaded`} />
    },
  },
]
