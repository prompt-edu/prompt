import { Card, CardContent, CardHeader, CardTitle } from '@tumaet/prompt-ui-components'
import { FileText } from 'lucide-react'
import type { ValidationResult } from '../../../../interfaces/validationResult'

interface DataCompletionSummaryProps {
  checks: ValidationResult[]
}

export const DataCompletionSummary = ({ checks }: DataCompletionSummaryProps) => {
  const total = checks.length
  const completed = checks.filter((check) => check.isValid).length

  return (
    <Card className='border-amber-100'>
      <CardHeader className='pb-2'>
        <CardTitle className='text-sm font-medium text-amber-700 flex items-center gap-2'>
          <FileText className='h-4 w-4' />
          Data from Previous Phases
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className='text-2xl font-bold'>
          {total > 0 ? Math.round((completed / total) * 100) : 0}%
        </div>
        <p className='text-xs text-muted-foreground'>
          {completed} of {total} checks complete
        </p>
      </CardContent>
    </Card>
  )
}
