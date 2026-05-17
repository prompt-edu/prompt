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
          <DialogTitle>Send reminder?</DialogTitle>
          <DialogDescription>
            {confirmationReminderType &&
              `This sends ${confirmationReminderType.label.toLowerCase()} reminders to students who still have incomplete evaluations.`}
            {previousSentAt &&
              ` A reminder for this type was already sent on ${formatSentAt(previousSentAt)}.`}
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
