import type { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { zodResolver } from '@hookform/resolvers/zod'
import { Form, FormField, FormItem, FormMessage } from '@tumaet/prompt-ui-components'
import type React from 'react'
import { forwardRef, useImperativeHandle } from 'react'
import { useForm } from 'react-hook-form'
import { checkCheckBoxQuestion } from '../../utils/CheckBoxRequirements'
import type { QuestionMultiSelectFormRef } from '../../utils/QuestionMultiSelectFormRef'
import { CheckboxQuestion } from './CheckboxQuestion'
import { MultiSelectQuestion } from './MultiSelectQuestion'
import { createValidationSchema } from './validationSchema'

interface ApplicationQuestionMultiSelectFormProps {
  question: ApplicationQuestionMultiSelect
  initialAnswers?: string[]
  isInstructorView?: boolean
}

export const ApplicationQuestionMultiSelectForm = forwardRef(
  function ApplicationQuestionMultiSelectFormInner(
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
