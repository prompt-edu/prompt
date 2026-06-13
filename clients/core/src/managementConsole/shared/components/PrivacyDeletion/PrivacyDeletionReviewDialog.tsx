import {
  type AdminPrivacyDeletionRequest,
  type AuditorDecision,
  decideOnDeletionRequest,
} from '@core/network/queries/privacyStudentDataDeletion'
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
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Mail, Recycle, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { PrivacyServiceAvailability } from '../Privacy/PrivacyServiceAvailability'

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
    <Dialog open={!!request} onOpenChange={handleOpenChange}>
      <DialogContent className='max-w-4xl'>
        <DialogHeader>
          <DialogTitle>Review Deletion Request</DialogTitle>
        </DialogHeader>

        {request && (
          <div className='grid gap-6 lg:grid-cols-2 text-sm'>
            <div className='flex flex-col gap-4'>
              <div>
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
                <p className='text-muted-foreground mt-2'>
                  on {new Date(request.requested_at).toLocaleString()}
                </p>
                <p>would like to delete all it&apos;s personal data</p>
              </div>

              <hr className='border-border' />

              <div className='flex flex-col gap-3'>
                <p className='font-medium'>How it works</p>
                <ul className='flex flex-col gap-2.5'>
                  <li className='flex items-start gap-3'>
                    <Trash2 className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                    <span>
                      Approval starts the deletion immediately across all services. This action
                      cannot be undone.
                    </span>
                  </li>
                  <li className='flex items-start gap-3'>
                    <Recycle className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                    <span>Rejection leaves the user&apos;s data unchanged.</span>
                  </li>
                  <li className='flex items-start gap-3'>
                    <Mail className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                    <span>
                      The user is notified of your decision and may submit another deletion request
                      later.
                    </span>
                  </li>
                </ul>

                <PrivacyServiceAvailability />
              </div>
            </div>

            <div className='flex flex-col gap-4'>
              <div className='flex flex-col gap-3'>
                <p className='font-medium'>What gets deleted</p>
                <p className='text-muted-foreground'>
                  Approval permanently removes the user&apos;s personal data. Only services the user
                  has actually interacted with are contacted, the rest are skipped. This may include
                  (depending on the user&apos;s history):
                </p>
                <ul className='list-disc pl-5 text-muted-foreground space-y-1'>
                  <li>Course enrollments</li>
                  <li>Application data</li>
                  <li>Assessment results</li>
                  <li>Team allocation</li>
                  <li>Instructor notes</li>
                </ul>
              </div>

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
          <Button onClick={() => mutation.mutate('approve')} disabled={mutation.isPending}>
            Approve and Start Deletion
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
