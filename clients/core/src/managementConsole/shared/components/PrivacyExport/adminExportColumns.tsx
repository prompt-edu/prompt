import { type AdminExportDoc, type AdminPrivacyExport, ExportStatus } from '@core/network/queries/privacyStudentDataExport'
import { Tooltip, TooltipContent, TooltipTrigger } from '@tumaet/prompt-ui-components'
import { ColumnDef } from '@tanstack/react-table'

function CountWithTooltip({ docs, label }: { docs: AdminExportDoc[]; label: string }) {
  if (docs.length === 0) return <span className='text-muted-foreground'>0 {label}</span>
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className='cursor-default underline decoration-dotted'>
          {docs.length} {label}
        </span>
      </TooltipTrigger>
      <TooltipContent>
        <ul className='text-xs space-y-0.5'>
          {docs.map((d) => (
            <li key={d.source_name}>{d.source_name}</li>
          ))}
        </ul>
      </TooltipContent>
    </Tooltip>
  )
}

function DocSummaryCell({ docs }: { docs: AdminExportDoc[] }) {
  if (docs.length === 0) return <span className='text-muted-foreground'>no docs</span>

  const succeeded = docs.filter((d) => d.status === ExportStatus.complete)
  const noData = docs.filter((d) => d.status === ExportStatus.no_data)
  const failed = docs.filter((d) => d.status === ExportStatus.failed)
  const downloaded = succeeded.filter((d) => d.downloaded)

  return (
    <div className='flex flex-col gap-0.5 text-sm'>
      <span className='text-muted-foreground text-xs'>{docs.length} total</span>
      <div className='flex flex-wrap gap-x-3 gap-y-0.5'>
        <CountWithTooltip docs={succeeded} label='ok' />
        {noData.length > 0 && <CountWithTooltip docs={noData} label='no data' />}
        {failed.length > 0 && <CountWithTooltip docs={failed} label='failed' />}
        {succeeded.length > 0 && (
          <CountWithTooltip docs={downloaded} label={`/ ${succeeded.length} downloaded`} />
        )}
      </div>
    </div>
  )
}

export const adminExportColumns: ColumnDef<AdminPrivacyExport>[] = [
  {
    accessorKey: 'status',
    header: 'Status',
    cell: (info) => info.getValue(),
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
      return <span className={isExpired ? 'text-red-800' : ''}>{validUntil.toLocaleString()}</span>
    },
  },
  {
    id: 'documents',
    header: 'Documents',
    cell: ({ row }) => <DocSummaryCell docs={row.original.docs} />,
  },
]
