import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Input,
  Label,
} from '@tumaet/prompt-ui-components'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { useEffect, useState } from 'react'

interface DestructiveResetDialogProps {
  open: boolean
  title: string
  description: string
  actionLabel: string
  isPending?: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => Promise<void> | void
}

export const DestructiveResetDialog = ({
  open,
  title,
  description,
  actionLabel,
  isPending = false,
  onOpenChange,
  onConfirm,
}: DestructiveResetDialogProps) => {
  const [confirmation, setConfirmation] = useState('')

  useEffect(() => {
    if (!open) setConfirmation('')
  }, [open])

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className='flex items-center gap-2'>
            <AlertTriangle className='h-5 w-5 text-destructive' />
            {title}
          </AlertDialogTitle>
          <AlertDialogDescription className='space-y-3'>
            <span className='block'>{description}</span>
            <span className='block font-medium text-foreground'>
              Type <span className='font-mono'>RESET</span> to confirm.
            </span>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <div className='space-y-2'>
          <Label htmlFor='reset-confirmation'>Confirmation</Label>
          <Input
            id='reset-confirmation'
            autoComplete='off'
            value={confirmation}
            onChange={(event) => setConfirmation(event.target.value)}
            placeholder='RESET'
          />
        </div>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            disabled={confirmation !== 'RESET' || isPending}
            className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            onClick={() => void onConfirm()}
          >
            {isPending ? <Loader2 className='mr-2 h-4 w-4 animate-spin' /> : null}
            {actionLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
