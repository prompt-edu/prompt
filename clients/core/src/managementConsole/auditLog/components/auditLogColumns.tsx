import { Badge, type PromptTableColumnDef } from '@tumaet/prompt-ui-components'
import type { AuditEntry } from '../interfaces/auditLog'

export const getAuditLogColumns = (): PromptTableColumnDef<AuditEntry>[] => [
  {
    accessorKey: 'createdAt',
    header: 'Time',
    cell: ({ row }) => new Date(row.original.createdAt).toLocaleString(),
  },
  {
    accessorKey: 'actorName',
    header: 'Actor',
    cell: ({ row }) => (
      <div className='flex flex-col'>
        <span>{row.original.actorName || 'Unknown'}</span>
        <span className='text-xs text-muted-foreground'>{row.original.actorEmail}</span>
      </div>
    ),
  },
  {
    accessorKey: 'actorRole',
    header: 'Role',
  },
  {
    accessorKey: 'action',
    header: 'Action',
  },
  {
    accessorKey: 'outcome',
    header: 'Outcome',
    cell: ({ row }) => (
      <Badge variant={row.original.outcome === 'denied' ? 'destructive' : 'secondary'}>
        {row.original.outcome}
      </Badge>
    ),
  },
  {
    id: 'entity',
    header: 'Entity',
    accessorFn: (row) => row.entityName || row.entityID || '',
    cell: ({ row }) => row.original.entityName || row.original.entityID || '-',
  },
  {
    accessorKey: 'sourceService',
    header: 'Source',
  },
]
