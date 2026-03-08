import React, { forwardRef, useImperativeHandle } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Form, FormField, FormItem, FormMessage } from '@tumaet/prompt-ui-components'
import { QuestionMultiSelectFormRef } from '../../utils/QuestionMultiSelectFormRef'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { createValidationSchema } from './validationSchema'
import { CheckboxQuestion } from './CheckboxQuestion'
import { MultiSelectQuestion } from './MultiSelectQuestion'
import { checkCheckBoxQuestion } from '../../utils/CheckBoxRequirements'

interface ApplicationQuestionMultiSelectFormProps {
  question: ApplicationQuestionMultiSelect
  initialAnswers?: string[]
  isInstructorView?: boolean
}

export const ApplicationQuestionMultiSelectForm = forwardRef(
  function ApplicationQuestionMultiSelectForm(
    {
      question,
      initialAnswers = [],
      isInstructorView = false,
    }: ApplicationQuestionMultiSelectFormProps,
    ref: React.Ref<QuestionMultiSelectFormRef>,
  ) {
    const isCheckboxQuestion = checkCheckBoxQuestion(question)

    const validationSchema = createValidationSchema(question, isCheckboxQuestion)

    const form = useForm<{ answers: string[] }>({
      resolver: zodResolver(validationSchema),
      defaultValues: { answers: initialAnswers },
      mode: 'onChange',
    })

    useImperativeHandle(ref, () => ({
      async validate() {
        const isValid = await form.trigger()
        return isValid
      },
      getValues() {
        return { applicationQuestionID: question.id, answer: form.getValues().answers }
      },
    }))

    return (
      <Form {...form}>
        <form>
          <FormField
            control={form.control}
            name='answers'
            render={({ fieldState }) => (
              <FormItem>
                {isCheckboxQuestion ? (
                  <CheckboxQuestion
                    form={form}
                    question={question}
                    isInstructorView={isInstructorView}
                  />
                ) : (
                  <MultiSelectQuestion
                    form={form}
                    question={question}
                    initialAnswers={initialAnswers}
                    isInstructorView={isInstructorView}
                  />
                )}
                <FormMessage>{fieldState.error?.message}</FormMessage>
              </FormItem>
            )}
          />
        </form>
      </Form>
    )
  },
)
