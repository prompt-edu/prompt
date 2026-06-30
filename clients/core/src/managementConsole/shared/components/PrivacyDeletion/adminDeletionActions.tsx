import {
  type AdminPrivacyDeletionRequest,
  DeletionRequestStatus,
} from '@core/network/queries/privacyStudentDataDeletion'
import { RowAction } from '@tumaet/prompt-ui-components'
import { ShieldCheck } from 'lucide-react'

interface PassedDeps {
  onReview: (request: AdminPrivacyDeletionRequest) => void
}

export function getAdminDeletionActions({
  onReview,
}: PassedDeps): RowAction<AdminPrivacyDeletionRequest>[] {
  return [
    {
      label: 'Review',
      icon: <ShieldCheck className='w-4 h-4' />,
      onAction: (rows) => onReview(rows[0]),
      hide: (rows) =>
        rows.length !== 1 || rows[0].status !== DeletionRequestStatus.pending_approval,
    },
  ]
}
