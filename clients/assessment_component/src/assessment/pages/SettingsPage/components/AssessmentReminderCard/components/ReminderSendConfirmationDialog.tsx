import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@tumaet/prompt-ui-components'

import type { ReminderTypeConfig } from '../interfaces/ReminderTypeConfig'
import { formatSentAt } from '../utils'

interface ReminderSendConfirmationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  confirmationReminderType: ReminderTypeConfig | null
  previousSentAt?: string
  isSending: boolean
  onConfirm: () => void
}

export function ReminderSendConfirmationDialog({
  open,
  onOpenChange,
  confirmationReminderType,
  previousSentAt,
  isSending,
  onConfirm,
}: ReminderSendConfirmationDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Confirm Reminder Send</DialogTitle>
          <DialogDescription>
            {confirmationReminderType &&
              `Send reminder emails for ${confirmationReminderType.label.toLowerCase()} to ${confirmationReminderType.recipientCount} currently incomplete ${
                confirmationReminderType.recipientCount === 1 ? 'student' : 'students'
              }?`}
            {previousSentAt &&
              ` A reminder for this type was already sent on ${formatSentAt(previousSentAt)}. Sending again will resend to currently incomplete students.`}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onConfirm} disabled={isSending}>
            {isSending ? 'Sending...' : 'Confirm Send'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
