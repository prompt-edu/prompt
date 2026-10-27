import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@tumaet/prompt-ui-components'
import { AlertTriangle } from 'lucide-react'

interface ShowProfilePicturesConfirmationProps {
  isOpen: boolean
  setOpen: (value: boolean) => void
  onConfirm: () => void
}

export const ShowProfilePicturesConfirmation = ({
  isOpen,
  setOpen,
  onConfirm,
}: ShowProfilePicturesConfirmationProps) => (
  <AlertDialog open={isOpen} onOpenChange={setOpen}>
    <AlertDialogContent className='max-w-[425px] [&>*]:min-w-0'>
      <AlertDialogHeader>
        <AlertDialogTitle className='flex items-center gap-2'>
          <AlertTriangle className='h-5 w-5' />
          Show profile pictures during the review?
        </AlertDialogTitle>
        <AlertDialogDescription className='text-base'>
          Applicants who already applied were told that their picture is not shown to the people
          reviewing their application.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <div className='mt-4 rounded-md bg-muted p-4 text-sm text-muted-foreground'>
        Once you save, the pictures of all applicants who are not accepted yet become visible to
        reviewers, including the pictures of applicants who were told they are hidden.
      </div>
      <AlertDialogFooter className='mt-6 flex-col-reverse sm:flex-row'>
        <AlertDialogCancel className='mt-3 sm:mt-0'>Keep hidden</AlertDialogCancel>
        <AlertDialogAction onClick={onConfirm}>Show pictures</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
)
