import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import type { ReactNode } from 'react'
import { type ColumnTarget, TARGET_LABELS } from '../utils/matchImportColumns'

interface ImportMappingStepProps {
  headers: string[]
  mapping: Record<string, ColumnTarget>
  error: string | null
  onChange: (header: string, target: ColumnTarget) => void
}

const TARGET_ORDER: ColumnTarget[] = [
  'firstName',
  'lastName',
  'universityLogin',
  'email',
  'matriculationNumber',
  'gender',
  'nationality',
  'studyProgram',
  'studyDegree',
  'currentSemester',
  'question',
  'ignore',
]

export const ImportMappingStep = ({
  headers,
  mapping,
  error,
  onChange,
}: ImportMappingStepProps): ReactNode => {
  return (
    <div className='space-y-4'>
      <p className='text-sm text-muted-foreground'>
        Map each CSV column to a student attribute, import it as an application question, or ignore
        it. The four required fields must each be mapped to one column.
      </p>

      {error && <p className='text-sm text-destructive'>{error}</p>}

      <div className='max-h-[360px] overflow-auto space-y-2 pr-1'>
        {headers.map((header) => (
          <div key={header} className='grid grid-cols-2 items-center gap-3'>
            <span className='text-sm font-medium truncate' title={header}>
              {header}
            </span>
            <Select
              value={mapping[header]}
              onValueChange={(value) => onChange(header, value as ColumnTarget)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {TARGET_ORDER.map((target) => (
                  <SelectItem key={target} value={target}>
                    {TARGET_LABELS[target]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        ))}
      </div>
    </div>
  )
}
