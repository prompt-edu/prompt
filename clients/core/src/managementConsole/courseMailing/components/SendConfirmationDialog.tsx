import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@tumaet/prompt-ui-components'
import { Loader2, Send } from 'lucide-react'

interface SendConfirmationDialogProps {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => void
  recipientCount: number
  isPending?: boolean
}

export const SendConfirmationDialog = ({
  isOpen,
  onClose,
  onConfirm,
  recipientCount,
  isPending,
}: SendConfirmationDialogProps) => {
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Send this campaign?</DialogTitle>
          <DialogDescription>
            This will send the email to <strong>{recipientCount}</strong>{' '}
            {recipientCount === 1 ? 'recipient' : 'recipients'}. This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant='outline' onClick={onClose} disabled={isPending}>
            Cancel
          </Button>
          <Button onClick={onConfirm} disabled={isPending || recipientCount === 0}>
            {isPending ? (
              <>
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                Sending...
              </>
            ) : (
              <>
                <Send className='mr-2 h-4 w-4' />
                Send to {recipientCount}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
