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

interface AllowedLengthFieldProps {
  form: UseFormReturn<QuestionConfigFormData>
}

export const AllowedLengthField = ({ form }: AllowedLengthFieldProps) => {
  return (
    <FormField
      control={form.control}
      name='allowedLength'
      render={({ field }) => (
        <FormItem>
          <FormLabel>
            Allowed Length <span className='text-destructive'> *</span>
          </FormLabel>
          <FormControl>
            <Input
              type='number'
              {...field}
              min={1}
              onChange={(e) => {
                const value = e.target.value
                field.onChange(value === '' ? '' : Number(value))
              }}
            />
          </FormControl>
          <FormDescription>
            The allowed number of chars to be entered. Recommended default is 500 chars. For {'<'}{' '}
            100 chars, the input will be a single line. For {'>'}100 chars, the input will be a
            textarea.
          </FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
