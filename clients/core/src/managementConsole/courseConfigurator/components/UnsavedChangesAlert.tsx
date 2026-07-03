import { Alert, AlertDescription, AlertTitle, Button } from '@tumaet/prompt-ui-components'
import { AlertCircle } from 'lucide-react'

interface UnsavedChangesAlertProps {
  handleRevert: () => void
  saveChanges: () => void
}

export function UnsavedChangesAlert({ handleRevert, saveChanges }: UnsavedChangesAlertProps) {
  return (
    <Alert variant='destructive' className='shadow-lg bg-white dark:bg-[#09090B]'>
      <AlertCircle className='h-4 w-4' />
      <AlertTitle>Unsaved Changes</AlertTitle>
      <AlertDescription className='flex items-center justify-between'>
        <span className='text-black dark:text-white'>
          This board has been modified. Would you like to save your changes?
        </span>
        <div className='space-y-1'>
          <Button
            variant='outline'
            size='sm'
            onClick={handleRevert}
            className='w-full text-black dark:text-white'
          >
            Revert
          </Button>
          <Button variant='default' size='sm' onClick={saveChanges} className='w-full'>
            Save
          </Button>
        </div>
      </AlertDescription>
    </Alert>
  )
}
