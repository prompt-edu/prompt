import { FormControl, FormField, FormItem, FormLabel, Input } from '@tumaet/prompt-ui-components'
import { UseFormReturn } from 'react-hook-form'
import { QuestionConfigFormData } from '@core/validations/questionConfig'

interface AllowedFileTypesFieldProps {
  form: UseFormReturn<QuestionConfigFormData>
}

export const AllowedFileTypesField = ({ form }: AllowedFileTypesFieldProps) => {
  return (
    <FormField
      control={form.control}
      name='allowedFileTypes'
      render={({ field }) => (
        <FormItem>
          <FormLabel>Allowed File Types</FormLabel>
          <FormControl>
            <Input
              placeholder='e.g., .pdf,.doc,.docx or application/pdf'
              {...field}
              value={field.value || ''}
            />
          </FormControl>
        </FormItem>
      )}
    />
  )
}
