import { QuestionConfigFormData } from '@core/validations/questionConfig'
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
} from '@tumaet/prompt-ui-components'
import { UseFormReturn } from 'react-hook-form'

interface ValidationRegexFieldProps {
  form: UseFormReturn<QuestionConfigFormData>
}

export const ValidationRegexField = ({ form }: ValidationRegexFieldProps) => {
  return (
    <FormField
      control={form.control}
      name='validationRegex'
      render={({ field }) => (
        <FormItem>
          <FormLabel>Validation Regex</FormLabel>
          <FormControl>
            <Input {...field} placeholder='Enter validation regex' />
          </FormControl>
          <FormDescription>
            The regex will be used as a validation rule for the input. Leave empty for no
            validation.
          </FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
