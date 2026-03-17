import { FormControl, FormField, FormItem, FormLabel, Input } from '@tumaet/prompt-ui-components'
import { UseFormReturn } from 'react-hook-form'
import { QuestionConfigFormData } from '@core/validations/questionConfig'

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
              onChange={(e) => field.onChange(e.target.value ? parseInt(e.target.value) : '')}
            />
          </FormControl>
        </FormItem>
      )}
    />
  )
}
