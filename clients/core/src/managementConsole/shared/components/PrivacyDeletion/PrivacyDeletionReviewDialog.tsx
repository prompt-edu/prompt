import {
  type AdminPrivacyDeletionRequest,
  type AuditorDecision,
  decideOnDeletionRequest,
} from '@core/network/queries/privacyStudentDataDeletion'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  ProfilePicture,
  Textarea,
} from '@tumaet/prompt-ui-components'
import { AlertTriangle } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  PrivacyDeletionHowItWorks,
  PrivacyDeletionWhatGetsDeleted,
} from './PrivacyDeletionExplainerContent'

const APPROVE_WAIT_SECONDS = 5

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
  const [approveRemaining, setApproveRemaining] = useState(APPROVE_WAIT_SECONDS)

  useEffect(() => {
    if (!request) return
    setApproveRemaining(APPROVE_WAIT_SECONDS)
    const interval = setInterval(() => {
      setApproveRemaining((prev) => {
        if (prev <= 1) {
          clearInterval(interval)
          return 0
        }
        return prev - 1
      })
    }, 1000)
    return () => clearInterval(interval)
  }, [request?.id])

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
    <Dialog open={!!request} onOpenChange={handleOpenChange}>
      <DialogContent className='max-w-4xl'>
        <DialogHeader>
          <DialogTitle>Review Deletion Request</DialogTitle>
        </DialogHeader>

        {request && (
          <div className='grid gap-6 lg:grid-cols-2 text-sm'>
            <div className='flex flex-col gap-4'>
              <div>
                {request.student_id && (
                  <Link
                    to={`/management/students/${request.student_id}`}
                    className='flex items-center gap-3 hover:text-blue-500'
                  >
                    <ProfilePicture
                      email={request.student_email ?? ''}
                      firstName={request.student_first_name ?? ''}
                      lastName={request.student_last_name ?? ''}
                      size='md'
                    />
                    <div className='flex flex-col'>
                      <span className='font-medium'>
                        {request.student_first_name} {request.student_last_name}
                      </span>
                      {request.student_email && (
                        <span className='text-muted-foreground'>{request.student_email}</span>
                      )}
                    </div>
                  </Link>
                )}
                {!request.student_id && (
                  <div className='rounded-lg border border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-900/30 p-4 flex items-start gap-3'>
                    <AlertTriangle className='h-5 w-5 shrink-0 text-amber-600 dark:text-amber-400 mt-0.5' />
                    <div>
                      <p className='font-semibold text-amber-900 dark:text-amber-200'>
                        Non-student requester
                      </p>
                      <p className='text-sm text-amber-700 dark:text-amber-300'>
                        This request has no associated student profile. The requester is likely an
                        instructor or other staff member. Deleting an instructor may have bigger
                        implications
                      </p>
                    </div>
                  </div>
                )}
                <p className='text-muted-foreground mt-2'>
                  on {new Date(request.requested_at).toLocaleString()}
                </p>
              </div>

              <hr className='border-border' />

              <PrivacyDeletionHowItWorks variant='review' />
            </div>

            <div className='flex flex-col gap-4'>
              <PrivacyDeletionWhatGetsDeleted variant='review' />

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
          </div>
        )}

        <DialogFooter>
          <Button variant='outline' onClick={onClose} disabled={mutation.isPending}>
            Cancel
          </Button>
          <Button onClick={() => mutation.mutate('reject')} disabled={mutation.isPending}>
            Reject
          </Button>
          <Button
            variant='destructive'
            onClick={() => mutation.mutate('approve')}
            disabled={mutation.isPending || approveRemaining > 0}
          >
            {approveRemaining > 0
              ? `Approve and Start Deletion (${approveRemaining}s)`
              : 'Approve and Start Deletion'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
