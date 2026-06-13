import {
  DeletionRequestStatus,
  getLatestStudentDataDeletion,
  getStudentDataDeletionStatus,
  requestStudentDataDeletion,
  type LatestDeletionResponse,
} from '@core/network/queries/privacyStudentDataDeletion'
import { Button, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import { PrivacyDeletionConfirmationDialog } from '../shared/components/PrivacyDeletion/PrivacyDeletionConfirmDialog'
import { PrivacyDeletionBanner } from '../shared/components/PrivacyDeletion/PrivacyDeletionBanner'
import { PrivacyDeletionSubrequestList } from '../shared/components/PrivacyDeletion/PrivacyDeletionSubrequestList'
import { usePrivacyRequestFlow } from '../shared/hooks/usePrivacyRequestFlow'

function isEndState(status: DeletionRequestStatus): boolean {
  return [
    DeletionRequestStatus.succeeded,
    DeletionRequestStatus.failed,
    DeletionRequestStatus.rejected,
  ].includes(status)
}

function getRequestIDFromLatest(latest: LatestDeletionResponse | undefined): string | undefined {
  if (latest?.status === 'exists') return latest.request.id
  return undefined
}

export function PrivacyDataDeletionPage() {
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false)

  const requestFlow = usePrivacyRequestFlow({
    resource: 'data-deletion',
    getLatest: getLatestStudentDataDeletion,
    extractIdFromLatest: getRequestIDFromLatest,
    createRequest: requestStudentDataDeletion,
    getStatus: getStudentDataDeletionStatus,
    isEndState,
  })

  const request = requestFlow.record
  const showRequestUI = !request || isEndState(request.status)

  const handleConfirm = () => {
    setConfirmDialogOpen(false)
    requestFlow.triggerCreate()
  }

  return (
    <div>
      <ManagementPageHeader>Data Deletion</ManagementPageHeader>

      {request && (
        <div className='mt-2 space-y-4'>
          <PrivacyDeletionBanner request={request} />
          <PrivacyDeletionSubrequestList subrequests={request.subrequests} />
        </div>
      )}

      {showRequestUI && (
        <div className={request ? 'mt-8' : ''}>
          <p className='text-muted-foreground'>
            Request the deletion of your personal data from our systems. An administrator must
            approve the request before the deletion is carried out. This action cannot be undone.
          </p>
          <div className='mt-6 flex flex-col items-start gap-2'>
            <Button disabled={requestFlow.isCreating} onClick={() => setConfirmDialogOpen(true)}>
              {request ? 'Request data deletion again' : 'Request data deletion'}
            </Button>
          </div>
        </div>
      )}

      <PrivacyDeletionConfirmationDialog
        isOpen={confirmDialogOpen}
        setIsOpen={setConfirmDialogOpen}
        handleConfirm={handleConfirm}
      />
    </div>
  )
}
