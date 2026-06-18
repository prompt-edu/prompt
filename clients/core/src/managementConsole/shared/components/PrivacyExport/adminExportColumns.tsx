import {
  type AdminExportDoc,
  type AdminPrivacyExport,
  ExportStatus,
} from '@core/network/queries/privacyStudentDataExport'
import { ColumnDef } from '@tanstack/react-table'
import { Archive, CircleCheck, CircleX, Info, Loader2 } from 'lucide-react'
import { HoverInfoText } from '../Privacy/HoverInfoText'
import { PrivacyStatusBadge } from '../Privacy/PrivacyStatusBadge'
import { RequesterCell, requesterAccessor } from '../Privacy/RequesterCell'

function CountWithTooltip({ docs, label }: { docs: AdminExportDoc[]; label: string }) {
  if (docs.length === 0) return <span>0 {label}</span>
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

const exportStatusColor: Record<ExportStatus, string> = {
  [ExportStatus.pending]: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
  [ExportStatus.complete]: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
  [ExportStatus.no_data]: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  [ExportStatus.failed]: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
  [ExportStatus.archived]: 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-300',
}

function exportStatusIcon(status: ExportStatus) {
  const className = 'h-3.5 w-3.5'
  switch (status) {
    case ExportStatus.pending:
      return <Loader2 className={`${className} animate-spin`} />
    case ExportStatus.complete:
      return <CircleCheck className={className} />
    case ExportStatus.no_data:
      return <Info className={className} />
    case ExportStatus.failed:
      return <CircleX className={className} />
    case ExportStatus.archived:
      return <Archive className={className} />
  }
}

function ExportStatusBadge({ status }: { status: ExportStatus }) {
  return (
    <PrivacyStatusBadge
      label={exportStatusLabel[status]}
      icon={exportStatusIcon(status)}
      colorClass={exportStatusColor[status]}
    />
  )
}

export const adminExportColumns: ColumnDef<AdminPrivacyExport>[] = [
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => <ExportStatusBadge status={row.original.status} />,
  },
  {
    id: 'requester',
    header: 'Requester',
    accessorFn: requesterAccessor,
    cell: ({ row }) => <RequesterCell {...row.original} />,
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
