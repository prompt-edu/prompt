import { Button } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'

interface PrivacyExportTriggerProps {
  progressOngoing: boolean
  rateLimited: boolean
  openDialog: () => void
}

export function PrivacyExportTrigger({
  progressOngoing,
  rateLimited,
  openDialog,
}: PrivacyExportTriggerProps) {
  return (
    <>
      <p className='text-muted-foreground'>
        Download a copy of all personal data stored about you in our systems.
      </p>
      <p className='mb-6 text-muted-foreground'>
        You must wait 30 days before requesting another export.
      </p>
      <Button disabled={progressOngoing || rateLimited} onClick={openDialog}>
        {progressOngoing && <Loader2 className='animate-spin mr-2 h-4 w-4' />}
        Request data export
      </Button>
    </>
  )
}
