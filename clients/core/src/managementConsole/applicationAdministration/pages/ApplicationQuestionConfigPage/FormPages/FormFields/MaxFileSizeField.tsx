import type { QuestionConfigFormData } from '@core/validations/questionConfig'
import { FormControl, FormField, FormItem, FormLabel, Input } from '@tumaet/prompt-ui-components'
import type { UseFormReturn } from 'react-hook-form'

interface MaxFileSizeFieldProps {
  form: UseFormReturn<QuestionConfigFormData>
}

export const MaxFileSizeField = ({ form }: MaxFileSizeFieldProps) => {
  return (
    <FormField
      control={form.control}
      name='maxFileSizeMB'
      render={({ field }) => (
        <FormItem>
          <FormLabel>Maximum File Size (MB)</FormLabel>
          <FormControl>
            <Input
              type='number'
              placeholder='10'
              {...field}
              value={field.value || ''}
              onChange={(e) => field.onChange(e.target.value ? parseInt(e.target.value, 10) : '')}
            />
          </FormControl>
        </FormItem>
      )}
    />
  )
}
