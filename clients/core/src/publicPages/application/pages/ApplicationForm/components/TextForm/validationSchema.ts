import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import * as z from 'zod'

export const createValidationSchema = (question: ApplicationQuestionText) =>
  z.object({
    answer: z
      .string()
      .min(question.isRequired ? 1 : 0, 'This field is required')
      .max(
        question.allowedLength || 255,
        `Answer must be less than ${question.allowedLength || 255} characters`,
      )
      .regex(
        new RegExp(question.validationRegex || '.*'),
        question.errorMessage || 'Invalid format',
      ),
  })
