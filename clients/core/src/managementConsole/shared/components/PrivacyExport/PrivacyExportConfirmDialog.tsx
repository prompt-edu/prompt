import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@tumaet/prompt-ui-components'
import { Clock, Download, Timer } from 'lucide-react'
import { PrivacyServiceAvailability } from '../Privacy/PrivacyServiceAvailability'

interface PrivacyExportConfirmationDialogProps {
  isOpen: boolean
  setIsOpen: (arg0: boolean) => void
  handleConfirm: () => void
}

export function PrivacyExportConfirmationDialog({
  isOpen,
  setIsOpen,
  handleConfirm,
}: PrivacyExportConfirmationDialogProps) {
  return (
    <AlertDialog open={isOpen} onOpenChange={setIsOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Request Data Export</AlertDialogTitle>
          <div className='flex flex-col gap-4 mt-2 text-sm'>
            <ul className='flex flex-col gap-2.5'>
              <li className='flex items-start gap-3'>
                <Clock className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                <span>You can request an export once every 30 days.</span>
              </li>
              <li className='flex items-start gap-3'>
                <Download className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                <span>
                  The export is available to download for 7 days, then permanently deleted.
                </span>
              </li>
              <li className='flex items-start gap-3'>
                <Timer className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                <span>The export process may take a few minutes.</span>
              </li>
            </ul>
            <PrivacyServiceAvailability forSelf />
          </div>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={handleConfirm}>Request Data Export</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
