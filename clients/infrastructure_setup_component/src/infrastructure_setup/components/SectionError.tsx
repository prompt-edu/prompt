import { Alert, AlertDescription, Button } from '@tumaet/prompt-ui-components'
import { AlertCircle } from 'lucide-react'

interface Props {
  message: string
  onRetry: () => void
}

// A failed load inside one part of a page, which leaves the rest of the page usable.
export const SectionError = ({ message, onRetry }: Props) => (
  <Alert variant='destructive'>
    <AlertCircle className='h-4 w-4' />
    <AlertDescription className='flex items-center justify-between gap-4'>
      <span>{message}</span>
      <Button variant='outline' size='sm' onClick={onRetry}>
        Retry
      </Button>
    </AlertDescription>
  </Alert>
)
