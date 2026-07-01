import { QuestionConfigFormData } from '@core/validations/questionConfig'
import { Checkbox, FormControl, FormField, FormItem, FormLabel } from '@tumaet/prompt-ui-components'
import { UseFormReturn } from 'react-hook-form'

interface RequiredFieldProps {
  form: UseFormReturn<QuestionConfigFormData>
}

export const RequiredField = ({ form }: RequiredFieldProps) => {
  return (
    <FormField
      control={form.control}
      name='isRequired'
      render={({ field }) => (
        <FormItem className='flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4'>
          <FormControl>
            <Checkbox checked={!!field.value} onCheckedChange={field.onChange} />
          </FormControl>
          <div className='space-y-1 leading-none'>
            <FormLabel>Question is required to be answered</FormLabel>
          </div>
        </FormItem>
      )}
    />
  )
}
