import {
  DeletionRequestStatus,
  DeletionSubrequestStatus,
  type PrivacyDeletionRequest,
} from '@core/network/queries/privacyStudentDataDeletion'
import { CircleCheck, CircleX, Loader2 } from 'lucide-react'
import { PrivacyStatusBanner } from '../Privacy/PrivacyStatusBanner'

interface PrivacyDeletionBannerProps {
  request: PrivacyDeletionRequest
}

type DisplayStatus = DeletionRequestStatus | 'partial'

function getDisplayStatus(request: PrivacyDeletionRequest): DisplayStatus {
  if (
    request.status === DeletionRequestStatus.failed &&
    request.subrequests.some((s) => s.status === DeletionSubrequestStatus.succeeded)
  ) {
    return 'partial'
  }
  return request.status
}

const requestStatusLabel: Record<DisplayStatus, string> = {
  [DeletionRequestStatus.pending_approval]: 'Waiting for admin approval',
  [DeletionRequestStatus.in_progress]: 'Deletion in progress',
  [DeletionRequestStatus.succeeded]: 'Deletion completed',
  [DeletionRequestStatus.failed]: 'Deletion failed',
  [DeletionRequestStatus.rejected]: 'Deletion request rejected',
  partial: 'Deletion succeeded with problems',
}

function RequestStatusIcon({ status }: { status: DisplayStatus }) {
  const className = 'h-5 w-5'
  switch (status) {
    case DeletionRequestStatus.pending_approval:
    case DeletionRequestStatus.in_progress:
      return <Loader2 className={`${className} animate-spin text-muted-foreground`} />
    case DeletionRequestStatus.succeeded:
      return <CircleCheck className={`${className} text-green-600 dark:text-green-400`} />
    case 'partial':
      return <CircleCheck className={`${className} text-amber-600 dark:text-amber-400`} />
    case DeletionRequestStatus.failed:
    case DeletionRequestStatus.rejected:
      return <CircleX className={`${className} text-red-600 dark:text-red-400`} />
  }
}

export function PrivacyDeletionBanner({ request }: PrivacyDeletionBannerProps) {
  const displayStatus = getDisplayStatus(request)
  const footer = request.auditor_note && (
    <div className='rounded border border-border bg-background px-3 py-2 text-sm text-foreground'>
      <p className='text-xs text-muted-foreground mb-1'>Note</p>
      <p className='whitespace-pre-wrap'>{request.auditor_note}</p>
    </div>
  )

  return (
    <PrivacyStatusBanner
      icon={<RequestStatusIcon status={displayStatus} />}
      title={requestStatusLabel[displayStatus]}
      meta={[
        `Requested on ${new Date(request.requested_at).toLocaleString()}`,
        request.auditor_responded_at &&
          `Reviewed by ${request.auditor_name || 'administrator'} on ${new Date(request.auditor_responded_at).toLocaleString()}`,
        request.completed_at && `Completed on ${new Date(request.completed_at).toLocaleString()}`,
      ]}
      footer={footer}
    />
  )
}
