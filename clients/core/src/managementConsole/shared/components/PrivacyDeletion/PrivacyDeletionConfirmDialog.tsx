import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@tumaet/prompt-ui-components'
import { Clock, ShieldAlert, Trash2 } from 'lucide-react'

interface PrivacyDeletionConfirmationDialogProps {
  isOpen: boolean
  setIsOpen: (open: boolean) => void
  handleConfirm: () => void
}

export function PrivacyDeletionConfirmationDialog({
  isOpen,
  setIsOpen,
  handleConfirm,
}: PrivacyDeletionConfirmationDialogProps) {
  return (
    <AlertDialog open={isOpen} onOpenChange={setIsOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Request Data Deletion</AlertDialogTitle>
          <div className='flex flex-col gap-4 mt-2 text-sm'>
            <ul className='flex flex-col gap-2.5'>
              <li className='flex items-start gap-3'>
                <ShieldAlert className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                <span>An administrator must approve your request before deletion starts.</span>
              </li>
              <li className='flex items-start gap-3'>
                <Trash2 className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                <span>
                  Once approved, your personal data will be permanently removed from our systems.
                </span>
              </li>
              <li className='flex items-start gap-3'>
                <Clock className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                <span>This action cannot be undone.</span>
              </li>
            </ul>
          </div>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={handleConfirm}>Request Deletion</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
