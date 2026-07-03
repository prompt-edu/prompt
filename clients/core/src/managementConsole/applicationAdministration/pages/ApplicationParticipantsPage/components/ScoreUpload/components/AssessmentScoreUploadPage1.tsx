import {
  Alert,
  AlertDescription,
  AlertTitle,
  Input,
  Label,
  Switch,
} from '@tumaet/prompt-ui-components'
import { AlertCircle } from 'lucide-react'
import { forwardRef, useImperativeHandle, useState } from 'react'

export interface Page1Ref {
  validate: () => boolean
  getValues: () => {
    scoreName: string
    hasThreshold: boolean
    threshold: string
  }
  reset: () => void
}

export const AssessmentScoreUploadPage1 = forwardRef<Page1Ref>(
  function AssessmentScoreUploadPage1Inner(_props, ref) {
    const [scoreName, setScoreName] = useState('')
    const [hasThreshold, setHasThreshold] = useState(false)
    const [threshold, setThreshold] = useState('')
    const [errors, setErrors] = useState<{ [key: string]: string }>({})

    const reset = () => {
      setScoreName('')
      setHasThreshold(false)
      setThreshold('')
      setErrors({})
    }

    const validate = () => {
      const newErrors: { [key: string]: string } = {}

      if (!scoreName.trim()) {
        newErrors.scoreName = 'Score name is required'
      }

      if (hasThreshold) {
        if (!threshold) {
          newErrors.threshold = 'Threshold is required when enabled'
        } else {
          const thresholdValue = parseFloat(threshold)
          if (Number.isNaN(thresholdValue)) {
            newErrors.threshold = 'Threshold must be a number'
          } else if (thresholdValue < 0) {
            newErrors.threshold = 'Threshold must be a positive number'
          }
        }
      }

      setErrors(newErrors)
      return Object.keys(newErrors).length === 0
    }

    useImperativeHandle(ref, () => ({
      validate,
      getValues: () => ({ scoreName, hasThreshold, threshold }),
      reset,
    }))

    return (
      <div className='space-y-6'>
        <div className='space-y-4'>
          <Label htmlFor='scoreName'>Name for the new score</Label>
          <Input
            id='scoreName'
            value={scoreName}
            onChange={(e) => setScoreName(e.target.value)}
            placeholder='e.g., Technical Challenge'
          />
          {errors.scoreName && <p className='text-sm text-red-500'>{errors.scoreName}</p>}
        </div>
        <div className='space-y-4'>
          <div className='flex items-center space-x-2'>
            <Switch id='threshold' checked={hasThreshold} onCheckedChange={setHasThreshold} />
            <Label htmlFor='threshold'>Set acceptance threshold</Label>
          </div>
          {hasThreshold && (
            <div className='space-y-2'>
              <Input
                type='number'
                value={threshold}
                onChange={(e) => setThreshold(e.target.value)}
                placeholder={
                  'Please enter the threshold as in your data. I.e., 50 or 0,5 depending on your data.'
                }
              />
              {errors.threshold && <p className='text-sm text-red-500'>{errors.threshold}</p>}
              <Alert>
                <AlertCircle className='h-4 w-4' />
                <AlertTitle>Note</AlertTitle>
                <AlertDescription>
                  Students with a score below this threshold will be automatically rejected.
                </AlertDescription>
              </Alert>
            </div>
          )}
        </div>
      </div>
    )
  },
)
