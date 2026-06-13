import {
  DeletionRequestStatus,
  type PrivacyDeletionRequest,
} from '@core/network/queries/privacyStudentDataDeletion'
import { CircleCheck, CircleX, Loader2 } from 'lucide-react'

interface PrivacyDeletionStatusCardProps {
  request: PrivacyDeletionRequest
}

const requestStatusLabel: Record<DeletionRequestStatus, string> = {
  [DeletionRequestStatus.pending_approval]: 'Waiting for admin approval',
  [DeletionRequestStatus.in_progress]: 'Deletion in progress',
  [DeletionRequestStatus.succeeded]: 'Deletion completed',
  [DeletionRequestStatus.failed]: 'Deletion failed',
  [DeletionRequestStatus.rejected]: 'Deletion request rejected',
}

function RequestStatusIcon({ status }: { status: DeletionRequestStatus }) {
  const className = 'h-5 w-5'
  switch (status) {
    case DeletionRequestStatus.pending_approval:
    case DeletionRequestStatus.in_progress:
      return <Loader2 className={`${className} animate-spin text-muted-foreground`} />
    case DeletionRequestStatus.succeeded:
      return <CircleCheck className={`${className} text-green-600 dark:text-green-400`} />
    case DeletionRequestStatus.failed:
    case DeletionRequestStatus.rejected:
      return <CircleX className={`${className} text-red-600 dark:text-red-400`} />
  }
}

export function PrivacyDeletionStatusCard({ request }: PrivacyDeletionStatusCardProps) {
  return (
    <div className='rounded-lg border border-border bg-muted p-4 flex flex-col gap-3'>
      <div className='flex items-center gap-3'>
        <div className='lg:mx-2'>
          <RequestStatusIcon status={request.status} />
        </div>
        <div className='flex-1'>
          <p className='font-semibold text-foreground'>{requestStatusLabel[request.status]}</p>
          <p className='text-xs text-muted-foreground mt-0.5'>
            Requested on {new Date(request.requested_at).toLocaleString()}
          </p>
          {request.auditor_responded_at && (
            <p className='text-xs text-muted-foreground mt-0.5'>
              Reviewed by {request.auditor_name || 'administrator'} on{' '}
              {new Date(request.auditor_responded_at).toLocaleString()}
            </p>
          )}
          {request.completed_at && (
            <p className='text-xs text-muted-foreground mt-0.5'>
              Completed on {new Date(request.completed_at).toLocaleString()}
            </p>
          )}
        </div>
      </div>

      {request.auditor_note && (
        <div className='rounded border border-border bg-background px-3 py-2 text-sm text-foreground'>
          <p className='text-xs text-muted-foreground mb-1'>Note</p>
          <p className='whitespace-pre-wrap'>{request.auditor_note}</p>
        </div>
      )}
    </div>
  )
}
