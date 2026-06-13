import {
  DeletionRequestStatus,
  getLatestStudentDataDeletion,
  getStudentDataDeletionStatus,
  requestStudentDataDeletion,
  type LatestDeletionResponse,
} from '@core/network/queries/privacyStudentDataDeletion'
import { useQuery } from '@tanstack/react-query'
import { Button, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'
import { PrivacyDeletionConfirmationDialog } from '../shared/components/PrivacyDeletion/PrivacyDeletionConfirmDialog'
import { PrivacyDeletionStatusCard } from '../shared/components/PrivacyDeletion/PrivacyDeletionStatusCard'
import { PrivacyDeletionSubrequestList } from '../shared/components/PrivacyDeletion/PrivacyDeletionSubrequestList'

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

  const latestQuery = useQuery({
    queryKey: ['privacy', 'data-deletion-latest'],
    queryFn: () => getLatestStudentDataDeletion(),
  })

  const requestQuery = useQuery({
    queryKey: ['privacy', 'data-deletion-create'],
    queryFn: () => requestStudentDataDeletion(),
    enabled: false,
  })

  const requestID = requestQuery.data?.id ?? getRequestIDFromLatest(latestQuery.data)

  const statusQuery = useQuery({
    queryKey: ['privacy', 'data-deletion-status', requestID],
    queryFn: () => getStudentDataDeletionStatus(requestID!),
    enabled: !!requestID,
    refetchInterval: (query) => {
      const status = query.state.data?.status
      if (status && isEndState(status)) return false
      return 3000
    },
  })

  const handleConfirm = () => {
    setConfirmDialogOpen(false)
    requestQuery.refetch()
  }

  const request = statusQuery.data
  const showRequestUI = !request || isEndState(request.status)

  return (
    <div>
      <ManagementPageHeader>Data Deletion</ManagementPageHeader>

      {request && (
        <div className='mt-2 space-y-4'>
          <PrivacyDeletionStatusCard request={request} />
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
            <Button disabled={requestQuery.isFetching} onClick={() => setConfirmDialogOpen(true)}>
              {requestQuery.isFetching && <Loader2 className='animate-spin mr-2 h-4 w-4' />}
              {request ? 'Request again' : 'Request data deletion'}
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
