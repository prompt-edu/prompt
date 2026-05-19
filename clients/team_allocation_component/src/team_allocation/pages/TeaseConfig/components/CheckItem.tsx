import { CheckCircle, AlertCircle } from 'lucide-react'
import { Card, CardContent } from '@tumaet/prompt-ui-components'
import type { ValidationResult } from '../../../interfaces/validationResult'

export const CheckItem = ({ check }: { check: ValidationResult }) => {
  return (
    <Card className={`border-l-4 ${check.isValid ? 'border-l-green-500' : 'border-l-yellow-500'}`}>
      <CardContent className='flex items-center gap-4 p-4'>
        <div className='shrink-0'>
          {check.isValid ? (
            <CheckCircle className='h-6 w-6 text-green-500' />
          ) : (
            <AlertCircle className='h-6 w-6 text-yellow-500' />
          )}
        </div>
        <div className='grow'>
          <div className='flex items-center justify-between'>
            <div className='flex items-center gap-2'>
              {check.icon}
              <h3 className='font-medium'>{check.label}</h3>
            </div>
          </div>

          {check.isValid ? (
            <p className='mt-2 text-sm text-green-600'>
              {check.details || 'All students have provided this information.'}
            </p>
          ) : (
            <div className='mt-2 space-y-1'>
              {check.problemDescription && (
                <p className='text-sm text-amber-700 font-medium'>{check.problemDescription}</p>
              )}
              {check.details && <p className='text-sm text-muted-foreground'>{check.details}</p>}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
