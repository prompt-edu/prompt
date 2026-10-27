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
          Show pictures to reviewers?
        </AlertDialogTitle>
        <AlertDialogDescription>
          Applicants who already applied were told their picture stays hidden until they are
          accepted. Once you save, reviewers see the pictures of all applicants who are not accepted
          yet.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Keep hidden</AlertDialogCancel>
        <AlertDialogAction onClick={onConfirm}>Show pictures</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
)
