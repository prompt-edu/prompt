import type { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import * as z from 'zod'

export const createValidationSchema = (
  question: ApplicationQuestionMultiSelect,
  isCheckboxQuestion: boolean,
) =>
  z.object({
    answers: z
      .array(z.string())
      .min(
        isCheckboxQuestion ? (question.isRequired ? 1 : 0) : question.minSelect,
        isCheckboxQuestion
          ? question.errorMessage || 'This checkbox is required'
          : question.maxSelect === 1
            ? 'Select an option'
            : `Select at least ${question.minSelect} option${question.minSelect > 1 ? 's' : ''}.`,
      )
      .max(
        question.maxSelect,
        `Select no more than ${question.maxSelect} option${question.maxSelect > 1 ? 's' : ''}.`,
      ),
  })
