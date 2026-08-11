import { PassStatus } from '@tumaet/prompt-shared-state'
import { Label, RadioGroup, RadioGroupItem } from '@tumaet/prompt-ui-components'
import { AlertTriangle } from 'lucide-react'
import type { ReactNode } from 'react'
import type { UnmatchedEnumValues } from '../utils/buildImportRequest'

interface ImportPreviewStepProps {
  totalRows: number
  newCount: number
  updateCount: number
  questionCount: number
  passStatus: PassStatus
  onPassStatusChange: (passStatus: PassStatus) => void
  unmatchedEnums: UnmatchedEnumValues
}

export const ImportPreviewStep = ({
  totalRows,
  newCount,
  updateCount,
  questionCount,
  passStatus,
  onPassStatusChange,
  unmatchedEnums,
}: ImportPreviewStepProps): ReactNode => {
  const enumWarnings = [
    { label: 'gender', values: unmatchedEnums.gender, fallback: 'prefer not to say' },
    { label: 'study degree', values: unmatchedEnums.studyDegree, fallback: 'bachelor' },
  ].filter((warning) => warning.values.length > 0)

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

      {enumWarnings.length > 0 && (
        <div className='flex gap-2 rounded-md border border-amber-500/50 bg-amber-50 p-3 text-sm text-amber-900 dark:bg-amber-950/40 dark:text-amber-200'>
          <AlertTriangle className='mt-0.5 h-4 w-4 shrink-0' />
          <div className='space-y-1'>
            <p className='font-medium'>Some values were not recognized and will use the default.</p>
            {enumWarnings.map((warning) => (
              <p key={warning.label}>
                Unrecognized {warning.label}: {warning.values.join(', ')} (stored as{' '}
                {warning.fallback}).
              </p>
            ))}
          </div>
        </div>
      )}

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
