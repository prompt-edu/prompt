import {
  type AdminPrivacyDeletionRequest,
  DeletionRequestStatus,
  DeletionSubrequestStatus,
  type PrivacyDeletionSubrequest,
} from '@core/network/queries/privacyStudentDataDeletion'
import { ColumnDef } from '@tanstack/react-table'
import { CircleCheck, CircleX, Clock, Loader2 } from 'lucide-react'
import { HoverInfoText } from '../Privacy/HoverInfoText'
import { RequesterDisplay } from '../Privacy/RequesterDisplay'

export const deletionRequestStatusLabel: Record<DeletionRequestStatus, string> = {
  [DeletionRequestStatus.pending_approval]: 'Pending approval',
  [DeletionRequestStatus.in_progress]: 'In progress',
  [DeletionRequestStatus.succeeded]: 'Completed',
  [DeletionRequestStatus.failed]: 'Failed',
  [DeletionRequestStatus.rejected]: 'Rejected',
}

const deletionRequestStatusStyles: Record<DeletionRequestStatus, string> = {
  [DeletionRequestStatus.pending_approval]:
    'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
  [DeletionRequestStatus.in_progress]:
    'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  [DeletionRequestStatus.succeeded]:
    'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
  [DeletionRequestStatus.failed]: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
  [DeletionRequestStatus.rejected]: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
}

function DeletionRequestStatusIcon({ status }: { status: DeletionRequestStatus }) {
  const className = 'h-3.5 w-3.5'
  switch (status) {
    case DeletionRequestStatus.pending_approval:
      return <Clock className={className} />
    case DeletionRequestStatus.in_progress:
      return <Loader2 className={`${className} animate-spin`} />
    case DeletionRequestStatus.succeeded:
      return <CircleCheck className={className} />
    case DeletionRequestStatus.failed:
    case DeletionRequestStatus.rejected:
      return <CircleX className={className} />
  }
}

function DeletionRequestStatusBadge({ status }: { status: DeletionRequestStatus }) {
  return (
    <span
      className={
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ' +
        deletionRequestStatusStyles[status]
      }
    >
      <DeletionRequestStatusIcon status={status} />
      {deletionRequestStatusLabel[status]}
    </span>
  )
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
    cell: ({ row }) => <DeletionRequestStatusBadge status={row.original.status} />,
  },
  {
    id: 'requester',
    header: 'Requester',
    accessorFn: (row) =>
      [row.student_first_name, row.student_last_name, row.student_email].filter(Boolean).join(' '),
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
