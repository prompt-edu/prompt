import { PassStatus } from '@tumaet/prompt-shared-state'
import { Label, RadioGroup, RadioGroupItem } from '@tumaet/prompt-ui-components'
import type { ReactNode } from 'react'

interface ImportPreviewStepProps {
  totalRows: number
  newCount: number
  updateCount: number
  questionCount: number
  passStatus: PassStatus
  onPassStatusChange: (passStatus: PassStatus) => void
}

export const ImportPreviewStep = ({
  totalRows,
  newCount,
  updateCount,
  questionCount,
  passStatus,
  onPassStatusChange,
}: ImportPreviewStepProps): ReactNode => {
  return (
    <div className='space-y-6'>
      <div className='grid grid-cols-3 gap-4'>
        <div className='rounded-md border p-4 text-center'>
          <p className='text-2xl font-semibold'>{newCount}</p>
          <p className='text-sm text-muted-foreground'>New students</p>
        </div>
        <div className='rounded-md border p-4 text-center'>
          <p className='text-2xl font-semibold'>{updateCount}</p>
          <p className='text-sm text-muted-foreground'>Existing (will update)</p>
        </div>
        <div className='rounded-md border p-4 text-center'>
          <p className='text-2xl font-semibold'>{questionCount}</p>
          <p className='text-sm text-muted-foreground'>New questions</p>
        </div>
      </div>

      <p className='text-sm text-muted-foreground'>
        {totalRows} student{totalRows === 1 ? '' : 's'} will be imported into this application
        phase.
      </p>

      <div className='space-y-2'>
        <p className='text-sm font-medium'>Status for imported students</p>
        <RadioGroup
          value={passStatus}
          onValueChange={(value) => onPassStatusChange(value as PassStatus)}
        >
          <div className='flex items-center gap-2'>
            <RadioGroupItem value={PassStatus.PASSED} id='import-status-passed' />
            <Label htmlFor='import-status-passed'>Accepted (passed)</Label>
          </div>
          <div className='flex items-center gap-2'>
            <RadioGroupItem value={PassStatus.NOT_ASSESSED} id='import-status-not-assessed' />
            <Label htmlFor='import-status-not-assessed'>Not assessed</Label>
          </div>
        </RadioGroup>
      </div>
    </div>
  )
}
