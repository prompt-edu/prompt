import { type AdminPrivacyExport } from '@core/network/queries/privacyStudentDataExport'
import { ColumnDef } from '@tanstack/react-table'
import { PrivacyExportStatus } from './PrivacyExportStatusBadge'

export const adminExportColumns: ColumnDef<AdminPrivacyExport>[] = [
  {
    id: 'status_icon',
    header: '',
    cell: ({ row }) => (
      <PrivacyExportStatus privacy_export_status={row.original.status} size={16} />
    ),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: 'userID',
    header: 'User ID',
    cell: (info) => info.getValue<string>().slice(0, 8) + '…',
  },
  {
    accessorKey: 'date_created',
    header: 'Requested',
    cell: (info) => new Date(info.getValue<string>()).toLocaleString(),
  },
  {
    accessorKey: 'valid_until',
    header: 'Valid Until',
    cell: ({ row }) => {
      const validUntil = new Date(row.original.valid_until)
      const isExpired = validUntil < new Date()
      return (
        <span className={isExpired ? 'text-muted-foreground line-through' : ''}>
          {validUntil.toLocaleString()}
        </span>
      )
    },
  },
  {
    id: 'downloaded',
    header: 'Downloaded',
    accessorFn: (row) => row.downloaded_docs,
    cell: ({ row }) => {
      const { downloaded_docs, total_docs, last_downloaded_at } = row.original
      if (total_docs === 0) return <span className='text-muted-foreground'>—</span>
      return (
        <div>
          <span>
            {downloaded_docs}/{total_docs}
          </span>
          {last_downloaded_at && (
            <p className='text-xs text-muted-foreground'>
              {new Date(last_downloaded_at).toLocaleDateString()}
            </p>
          )}
        </div>
      )
    },
  },
]
