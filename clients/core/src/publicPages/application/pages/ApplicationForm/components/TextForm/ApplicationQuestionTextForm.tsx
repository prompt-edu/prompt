import React, { forwardRef, useEffect, useImperativeHandle, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Textarea,
} from '@tumaet/prompt-ui-components'
import { QuestionTextFormRef } from '../../utils/QuestionTextFormRef'
import { createValidationSchema } from './validationSchema'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { FormDescriptionHTML } from '../FormDescriptionHTML'

interface ApplicationQuestionTextFormProps {
  question: ApplicationQuestionText
  initialAnswer?: string
  isInstructorView?: boolean
}

export const ApplicationQuestionTextForm = forwardRef(function ApplicationQuestionTextForm(
  { question, initialAnswer, isInstructorView = false }: ApplicationQuestionTextFormProps,
  ref: React.Ref<QuestionTextFormRef>,
) {
  const [charCount, setCharCount] = useState(initialAnswer?.length || 0)

  // Create validation schema dynamically based on question properties
  const validationSchema = createValidationSchema(question)

  const form = useForm<{ answer: string }>({
    resolver: zodResolver(validationSchema),
    defaultValues: { answer: initialAnswer ?? '' },
    mode: 'onChange',
  })

  // Expose validation method
  useImperativeHandle(ref, () => ({
    async validate() {
      const isValid = await form.trigger()
      return isValid
    },
    getValues() {
      return { applicationQuestionID: question.id, answer: form.getValues().answer }
    },
  }))

  // Watch for changes in the "answer" field to update character count
  useEffect(() => {
    const subscription = form.watch((value) => {
      setCharCount(value.answer?.length || 0)
    })
    return () => subscription.unsubscribe()
  }, [form, form.watch])

  const isTextArea = (question.allowedLength || 255) > 100

  return (
    <Form {...form}>
      <form>
        <FormField
          control={form.control}
          name='answer'
          render={({ field, fieldState }) => (
            <FormItem>
              <FormLabel>
                {question.title}
                {question.isRequired ? <span className='text-destructive'> *</span> : ''}
              </FormLabel>
              {!isInstructorView && question.description && (
                <FormDescriptionHTML htmlCode={question.description} />
              )}
              <FormControl>
                <div className='relative'>
                  {isTextArea ? (
                    <Textarea
                      {...field}
                      placeholder={question.placeholder || ''}
                      maxLength={question.allowedLength}
                      className='pr-12'
                      rows={isInstructorView ? 8 : 4}
                      disabled={isInstructorView}
                    />
                  ) : (
                    <Input
                      {...field}
                      placeholder={question.placeholder || ''}
                      maxLength={question.allowedLength}
                      className='pr-12'
                      disabled={isInstructorView}
                    />
                  )}
                  <div className='absolute right-2 bottom-2 text-sm text-gray-500'>
                    {charCount}/{question.allowedLength || 255}
                  </div>
                </div>
              </FormControl>
              <FormMessage>{fieldState.error?.message}</FormMessage>
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
})
