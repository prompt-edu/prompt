import {
  type AdminPrivacyDeletionRequest,
  type AuditorDecision,
  decideOnDeletionRequest,
} from '@core/network/queries/privacyStudentDataDeletion'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Button,
  Textarea,
} from '@tumaet/prompt-ui-components'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { AlertTriangle, ShieldAlert, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { PrivacyServiceAvailability } from '../Privacy/PrivacyServiceAvailability'
import { RequesterDisplay } from '../Privacy/RequesterDisplay'

interface PrivacyDeletionReviewDialogProps {
  request: AdminPrivacyDeletionRequest | null
  onClose: () => void
}

export function PrivacyDeletionReviewDialog({
  request,
  onClose,
}: PrivacyDeletionReviewDialogProps) {
  const queryClient = useQueryClient()
  const [note, setNote] = useState('')

  const mutation = useMutation({
    mutationFn: (decision: AuditorDecision) =>
      decideOnDeletionRequest(request!.id, { decision, note }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['privacy', 'admin', 'deletions'] })
      setNote('')
      onClose()
    },
  })

  const handleOpenChange = (open: boolean) => {
    if (!open && !mutation.isPending) {
      setNote('')
      onClose()
    }
  }

  return (
    <AlertDialog open={!!request} onOpenChange={handleOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Review Deletion Request</AlertDialogTitle>
        </AlertDialogHeader>

        {request && (
          <div className='flex flex-col gap-4 mt-2 text-sm'>
            <div className='rounded border border-border bg-muted px-3 py-2'>
              <p className='text-xs text-muted-foreground mb-1'>Requested by</p>
              <RequesterDisplay {...request} />
              <p className='text-xs text-muted-foreground mt-2'>
                on {new Date(request.requested_at).toLocaleString()}
              </p>
            </div>

            <ul className='flex flex-col gap-2.5'>
              <li className='flex items-start gap-3'>
                <Trash2 className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                <span>
                  Approval starts the deletion immediately across all registered services.
                </span>
              </li>
              <li className='flex items-start gap-3'>
                <ShieldAlert className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                <span>This action cannot be undone.</span>
              </li>
              <li className='flex items-start gap-3'>
                <AlertTriangle className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                <span>Rejection leaves the user&apos;s data unchanged.</span>
              </li>
            </ul>

            <PrivacyServiceAvailability />

            <div className='flex flex-col gap-1'>
              <label className='text-xs text-muted-foreground'>Note (optional)</label>
              <Textarea
                value={note}
                onChange={(e) => setNote(e.target.value)}
                disabled={mutation.isPending}
                placeholder='Let the requester know why'
                rows={3}
              />
            </div>
          </div>
        )}

        <AlertDialogFooter>
          <Button variant='outline' onClick={onClose} disabled={mutation.isPending}>
            Cancel
          </Button>
          <Button
            variant='destructive'
            onClick={() => mutation.mutate('reject')}
            disabled={mutation.isPending}
          >
            Reject
          </Button>
          <Button onClick={() => mutation.mutate('approve')} disabled={mutation.isPending}>
            Approve
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
