import { useState } from 'react'
import { UseFormReturn } from 'react-hook-form'
import {
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
  Input,
  Switch,
  TooltipProvider,
} from '@tumaet/prompt-ui-components'
import { DescriptionMinimalTiptapEditor } from '@tumaet/prompt-ui-components'
import { QuestionConfigFormData } from '@core/validations/questionConfig'

interface DescriptionFieldProps {
  form: UseFormReturn<QuestionConfigFormData>
  initialDescription?: string
}

export const DescriptionField = ({ form, initialDescription }: DescriptionFieldProps) => {
  // if a rich text is entered -> start with rich text editor
  const [useRichInput, setUseRichInput] = useState(initialDescription?.startsWith('<') ?? false)

  return (
    <FormField
      control={form.control}
      name='description'
      render={({ field }) => (
        <FormItem>
          <div className='flex items-center justify-between'>
            <FormLabel>Description</FormLabel>
            <div className='flex items-center space-x-2'>
              <Switch
                checked={useRichInput}
                onCheckedChange={setUseRichInput}
                aria-label='Toggle between standard and rich input'
              />
              <span className='text-sm text-muted-foreground'>Rich Text Editor</span>
            </div>
          </div>
          <FormControl>
            {useRichInput ? (
              <TooltipProvider>
                <DescriptionMinimalTiptapEditor
                  {...field}
                  className='w-full'
                  editorContentClassName='p-3'
                  output='html'
                  placeholder='Type your description here...'
                  autofocus={false}
                  editable={true}
                  editorClassName='focus:outline-hidden'
                />
              </TooltipProvider>
            ) : (
              <Input {...field} placeholder='Enter description text' />
            )}
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
