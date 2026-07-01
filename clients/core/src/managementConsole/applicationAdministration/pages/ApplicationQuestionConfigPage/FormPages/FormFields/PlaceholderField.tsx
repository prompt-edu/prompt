import { QuestionConfigFormData } from '@core/validations/questionConfig'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
} from '@tumaet/prompt-ui-components'
import { UseFormReturn } from 'react-hook-form'

interface PlaceholderFieldProps {
  form: UseFormReturn<QuestionConfigFormData>
}

export const PlaceholderField = ({ form }: PlaceholderFieldProps) => {
  return (
    <FormField
      control={form.control}
      name='placeholder'
      render={({ field }) => (
        <FormItem>
          <FormLabel>Placeholder</FormLabel>
          <FormControl>
            <Input {...field} placeholder='Enter placeholder text' />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
