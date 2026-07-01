import type { QuestionConfigFormData } from '@core/validations/questionConfig'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
} from '@tumaet/prompt-ui-components'
import type { UseFormReturn } from 'react-hook-form'

interface TitleFieldProps {
  form: UseFormReturn<QuestionConfigFormData>
}

export const TitleField = ({ form }: TitleFieldProps) => {
  return (
    <FormField
      control={form.control}
      name='title'
      render={({ field }) => (
        <FormItem>
          <FormLabel>
            Title <span className='text-destructive'> *</span>
          </FormLabel>
          <FormControl>
            <Input {...field} placeholder='Enter question title' />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
