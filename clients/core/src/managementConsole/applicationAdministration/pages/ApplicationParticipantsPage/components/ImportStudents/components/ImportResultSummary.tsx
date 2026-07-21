import type { ImportResult } from '@core/managementConsole/applicationAdministration/interfaces/import/importResult'
import { CheckCircle2 } from 'lucide-react'
import type { ReactNode } from 'react'

interface ImportResultSummaryProps {
  result: ImportResult
}

export const ImportResultSummary = ({ result }: ImportResultSummaryProps): ReactNode => {
  return (
    <div className='space-y-4'>
      <div className='flex items-center gap-2 text-green-600'>
        <CheckCircle2 className='h-5 w-5' />
        <span className='font-medium'>Import complete</span>
      </div>
      <div className='grid grid-cols-2 gap-4'>
        <div className='rounded-md border p-4 text-center'>
          <p className='text-2xl font-semibold'>{result.created}</p>
          <p className='text-sm text-muted-foreground'>Created</p>
        </div>
        <div className='rounded-md border p-4 text-center'>
          <p className='text-2xl font-semibold'>{result.updated}</p>
          <p className='text-sm text-muted-foreground'>Updated</p>
        </div>
      </div>
    </div>
  )
}
