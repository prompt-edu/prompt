import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from '@tumaet/prompt-ui-components'

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
          <AlertDialogDescription className='mt-2 space-y-1'>
            <p>You can only request a data export every 30 days.</p>
            <p>
              The export will be available for you to download for 7 days and permanently deleted
              after.
            </p>
            <p>The export process might take a few minutes.</p>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={handleConfirm}>Request Data Export</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
