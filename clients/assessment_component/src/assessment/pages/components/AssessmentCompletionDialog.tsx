import {
  Alert,
  AlertDescription,
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@tumaet/prompt-ui-components'

import { format } from 'date-fns'

interface AssessmentCompletionDialogProps {
  completed: boolean
  completedAt?: Date
  author?: string
  isPending: boolean
  dialogOpen: boolean
  setDialogOpen: (open: boolean) => void
  error: string | undefined
  setError: (error: string | undefined) => void
  handleConfirm: () => void
  isDeadlinePassed?: boolean
}

export function AssessmentCompletionDialog({
  completed,
  completedAt,
  author,
  isPending,
  dialogOpen,
  setDialogOpen,
  error,
  setError,
  handleConfirm,
  isDeadlinePassed = false,
}: AssessmentCompletionDialogProps) {
  const dialogDescription = completed ? (
    <>
      Marked as final {author ? <>by {author}</> : <></>}
      at {completedAt ? format(completedAt, 'MMM d, yyyy') : 'N/A'}
      <br /> Are you sure you want to reopen this for editing? This will allow you to make changes
      to your scores.
    </>
  ) : !isDeadlinePassed ? (
    <>
      <strong>Are you sure you want to mark this as final?</strong>
      <br />
      You can still unmark and edit it until the deadline. After the deadline, the form will be
      locked permanently and no further changes will be possible.
    </>
  ) : (
    <>
      <strong>Are you sure you want to mark this as final?</strong>
      <br />
      This will lock the form permanently and prevent further changes.
    </>
  )

  return (
    <div>
      <Dialog
        open={dialogOpen}
        onOpenChange={(open) => {
          if (!open) setError(undefined) // Clear error when closing dialog
          setDialogOpen(open)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{completed ? 'Reopen for Editing' : 'Mark as Final'}</DialogTitle>
            <DialogDescription>
              <DialogDescription>{dialogDescription}</DialogDescription>
            </DialogDescription>
          </DialogHeader>

          {error && (
            <Alert variant='destructive'>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <DialogFooter>
            <Button variant='outline' onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleConfirm} variant={completed ? 'destructive' : 'default'}>
              {isPending
                ? 'Processing...'
                : completed
                  ? 'Yes, Reopen for Editing'
                  : 'Yes, Mark as Final'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
