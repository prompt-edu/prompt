import {
  Button,
  cn,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@tumaet/prompt-ui-components'
import { AlertCircle, CheckCircle, Loader2, XCircle } from 'lucide-react'

interface ApplicationSavingDialogProps {
  showDialog: 'saving' | 'success' | 'error' | null
  onClose: () => void
  onNavigateBack: () => void
  navigateBackLabel?: string
  errorMessage?: string
  confirmationMailSent?: boolean
}

export const ApplicationSavingDialog = ({
  showDialog,
  onClose,
  onNavigateBack,
  navigateBackLabel = 'Back to Overview',
  errorMessage,
  confirmationMailSent = false,
}: ApplicationSavingDialogProps) => {
  const isOpen = showDialog !== null

  const getDialogContent = () => {
    switch (showDialog) {
      case 'saving':
        return {
          title: 'Saving Application',
          description: 'Please wait while we save your application...',
          icon: Loader2,
          iconColor: 'text-blue-500',
          iconBg: 'bg-blue-100',
        }
      case 'success':
        return {
          title: 'Application Saved',
          description: `Your application was successfully saved! ${
            confirmationMailSent
              ? 'We’ve sent you a confirmation mail, where you find all important information about the next steps.'
              : ''
          }`,
          icon: CheckCircle,
          iconColor: 'text-green-500',
          iconBg: 'bg-green-100',
        }
      case 'error':
        if (errorMessage?.includes('409')) {
          return {
            title: 'Registration Error',
            description:
              'This email is already registered, but with a different name. Please contact an instructor.',
            icon: AlertCircle,
            iconColor: 'text-yellow-500',
            iconBg: 'bg-yellow-100',
          }
        } else if (errorMessage?.includes('405')) {
          return {
            title: 'Duplicate Application',
            description:
              'There is already an application registered to this course with this email. ' +
              'Without a university account, you cannot modify your application after submission.',
            icon: XCircle,
            iconColor: 'text-red-500',
            iconBg: 'bg-red-100',
          }
        } else {
          return {
            title: 'Error',
            description: 'An error occurred while saving your application. Please try again later.',
            icon: XCircle,
            iconColor: 'text-red-500',
            iconBg: 'bg-red-100',
          }
        }
      default:
        return null
    }
  }

  const dialogContent = getDialogContent()

  if (!dialogContent) return <></>

  const IconComponent = dialogContent.icon

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) {
          onClose()
        }
      }}
    >
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <div className='flex flex-col items-center text-center'>
            <div
              className={cn(
                'flex h-20 w-20 items-center justify-center rounded-full',
                dialogContent.iconBg,
              )}
            >
              <IconComponent className={cn('h-10 w-10', dialogContent.iconColor)} />
            </div>
            <DialogTitle className='mt-4 text-xl font-semibold'>{dialogContent.title}</DialogTitle>
          </div>
          <DialogDescription className='text-center text-sm'>
            {dialogContent.description}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className='sm:justify-center'>
          {showDialog === 'saving' ? (
            <Button disabled className='w-full sm:w-auto'>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Please wait
            </Button>
          ) : showDialog === 'success' ? (
            <Button onClick={onNavigateBack} className='w-full sm:w-auto'>
              {navigateBackLabel}
            </Button>
          ) : (
            <Button variant='outline' onClick={onClose} className='w-full sm:w-auto'>
              Back
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
