import {
  type AdminPrivacyDeletionRequest,
  DeletionRequestStatus,
  DeletionSubrequestStatus,
  type PrivacyDeletionSubrequest,
} from '@core/network/queries/privacyStudentDataDeletion'
import { ColumnDef } from '@tanstack/react-table'
import { HoverInfoText } from '../Privacy/HoverInfoText'
import { RequesterDisplay } from '../Privacy/RequesterDisplay'

export const deletionRequestStatusLabel: Record<DeletionRequestStatus, string> = {
  [DeletionRequestStatus.pending_approval]: 'Pending approval',
  [DeletionRequestStatus.in_progress]: 'In progress',
  [DeletionRequestStatus.succeeded]: 'Completed',
  [DeletionRequestStatus.failed]: 'Failed',
  [DeletionRequestStatus.rejected]: 'Rejected',
}

function CountWithTooltip({ subs, label }: { subs: PrivacyDeletionSubrequest[]; label: string }) {
  if (subs.length === 0) return <span className='text-muted-foreground'>0 {label}</span>
  return (
    <HoverInfoText
      content={
        <ul className='text-xs space-y-0.5'>
          {subs.map((s) => (
            <li key={s.id}>{s.source_name}</li>
          ))}
        </ul>
      }
    >
      {subs.length} {label}
    </HoverInfoText>
  )
}

function SourceSummaryCell({ subs }: { subs: PrivacyDeletionSubrequest[] }) {
  if (subs.length === 0) return <span className='text-muted-foreground'>—</span>

  const succeeded = subs.filter((s) => s.status === DeletionSubrequestStatus.succeeded)
  const inProgress = subs.filter((s) => s.status === DeletionSubrequestStatus.in_progress)
  const pending = subs.filter((s) => s.status === DeletionSubrequestStatus.pending)
  const failed = subs.filter((s) => s.status === DeletionSubrequestStatus.failed)

  return (
    <div className='flex flex-col gap-0.5 text-sm'>
      <span className='text-muted-foreground text-xs'>{subs.length} total</span>
      <div className='flex flex-wrap gap-x-3 gap-y-0.5'>
        <CountWithTooltip subs={succeeded} label='ok' />
        {inProgress.length > 0 && <CountWithTooltip subs={inProgress} label='in progress' />}
        {pending.length > 0 && <CountWithTooltip subs={pending} label='pending' />}
        {failed.length > 0 && <CountWithTooltip subs={failed} label='failed' />}
      </div>
    </div>
  )
}

export const adminDeletionColumns: ColumnDef<AdminPrivacyDeletionRequest>[] = [
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => deletionRequestStatusLabel[row.original.status],
  },
  {
    id: 'requester',
    header: 'Requester',
    cell: ({ row }) => <RequesterDisplay {...row.original} />,
  },
  {
    accessorKey: 'requested_at',
    header: 'Requested',
    cell: (info) => new Date(info.getValue<string>()).toLocaleString(),
  },
  {
    id: 'reviewed',
    header: 'Reviewed',
    cell: ({ row }) => {
      const { auditor_name, auditor_email, auditor_responded_at, auditor_note } = row.original
      if (!auditor_responded_at) return <span className='text-muted-foreground text-sm'>—</span>
      const name = auditor_name || auditor_email || 'administrator'
      return (
        <div className='flex flex-col text-sm'>
          <span>{name}</span>
          <span className='text-muted-foreground text-xs'>
            on {new Date(auditor_responded_at).toLocaleString()}
          </span>
          {auditor_note && (
            <span className='text-xs mt-0.5'>
              <HoverInfoText
                content={<p className='whitespace-pre-wrap max-w-xs'>{auditor_note}</p>}
              >
                Note
              </HoverInfoText>
            </span>
          )}
        </div>
      )
    },
  },
  {
    accessorKey: 'completed_at',
    header: 'Completed',
    cell: ({ row }) => {
      const completedAt = row.original.completed_at
      if (!completedAt) return <span className='text-muted-foreground text-sm'>—</span>
      return <span className='text-sm'>{new Date(completedAt).toLocaleString()}</span>
    },
  },
  {
    id: 'sources',
    header: 'Sources',
    cell: ({ row }) => <SourceSummaryCell subs={row.original.subrequests} />,
  },
]
