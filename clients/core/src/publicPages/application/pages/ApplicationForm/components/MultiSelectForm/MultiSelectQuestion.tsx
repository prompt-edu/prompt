import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import {
  FormLabel,
  MultiSelect,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import React from 'react'
import { UseFormReturn } from 'react-hook-form'
import { FormDescriptionHTML } from '../FormDescriptionHTML'

interface MultiSelectQuestionProps {
  form: UseFormReturn<{ answers: string[] }>
  question: ApplicationQuestionMultiSelect
  initialAnswers: string[]
  isInstructorView?: boolean
}

export const MultiSelectQuestion: React.FC<MultiSelectQuestionProps> = ({
  form,
  question,
  initialAnswers,
  isInstructorView = false,
}) => {
  const multiSelectOptions = question.options.map((option) => ({
    label: option,
    value: option,
  }))

  return (
    <>
      <FormLabel>
        {question.title}
        {question.minSelect > 0 && <span className='text-destructive'> *</span>}
      </FormLabel>
      {!isInstructorView && question.description && (
        <FormDescriptionHTML htmlCode={question.description} />
      )}
      {question.maxSelect > 1 ? (
        <MultiSelect
          options={multiSelectOptions}
          placeholder={question.placeholder || 'Please select...'}
          defaultValue={initialAnswers}
          onValueChange={(values) => {
            form.setValue('answers', values, { shouldValidate: true })
          }}
          maxCount={question.maxSelect}
          variant='inverted'
          disabled={isInstructorView}
        />
      ) : (
        <Select
          onValueChange={(value) => {
            form.setValue('answers', [value], { shouldValidate: true })
          }}
          defaultValue={initialAnswers.length === 1 ? initialAnswers[0] : ''}
          disabled={isInstructorView}
        >
          <SelectTrigger>
            <SelectValue placeholder={question.placeholder || 'Please select...'} />
          </SelectTrigger>
          <SelectContent>
            {question.options.map((option) => (
              <SelectItem key={option} value={option}>
                {option}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}
    </>
  )
}
