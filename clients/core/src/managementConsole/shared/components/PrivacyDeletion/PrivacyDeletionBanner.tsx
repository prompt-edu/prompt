import {
  DeletionRequestStatus,
  DeletionSubrequestStatus,
  type PrivacyDeletionRequest,
} from '@core/network/queries/privacyStudentDataDeletion'
import {
  PrivacyStatusBanner,
  type PrivacyStatusBannerState,
} from '../Privacy/PrivacyStatusBanner'

interface PrivacyDeletionBannerProps {
  request: PrivacyDeletionRequest
}

function getState(request: PrivacyDeletionRequest): PrivacyStatusBannerState {
  switch (request.status) {
    case DeletionRequestStatus.pending_approval:
      return 'pending_approval'
    case DeletionRequestStatus.in_progress:
      return 'in_progress'
    case DeletionRequestStatus.succeeded:
      return 'success'
    case DeletionRequestStatus.failed: {
      const anySucceeded = request.subrequests.some(
        (s) => s.status === DeletionSubrequestStatus.succeeded,
      )
      return anySucceeded ? 'partial' : 'failure'
    }
    case DeletionRequestStatus.rejected:
      return 'rejected'
  }
}

export function PrivacyDeletionBanner({ request }: PrivacyDeletionBannerProps) {
  const state = getState(request)

  const footer = request.auditor_note && (
    <div className='rounded border border-border bg-background px-3 py-2 text-sm text-foreground'>
      <p className='text-xs text-muted-foreground mb-1'>Note</p>
      <p className='whitespace-pre-wrap'>{request.auditor_note}</p>
    </div>
  )

  return (
    <PrivacyStatusBanner
      subject='Deletion'
      state={state}
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
