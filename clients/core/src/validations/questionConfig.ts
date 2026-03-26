// validations/questionConfig.ts
import * as z from 'zod'

// Base schema containing common fields and the discriminant 'type'
const baseQuestionSchema = z.object({
  type: z.enum(['text', 'multi-select', 'file-upload']), // Discriminant field
  title: z.string().min(1, 'Title is required'),
  description: z.string().optional(),
  placeholder: z.string().optional(),
  errorMessage: z.string().optional(),
  isRequired: z.boolean(),
  accessibleForOtherPhases: z.boolean(),
  accessKey: z.string().optional(),
})

// Schema for text questions
export const textQuestionSchema = baseQuestionSchema.extend({
  type: z.literal('text'), // Ensure the type is 'text'
  validationRegex: z
    .string()
    .optional()
    .refine(
      (val) => {
        if (!val) return true // If empty, consider it valid
        try {
          new RegExp(val)
          return true
        } catch {
          return false
        }
      },
      {
        message: 'Invalid regex pattern',
      },
    ),
  allowedLength: z.number().min(1, 'Allowed length must be at least 1'),
})

// Schema for multi-select questions
export const multiSelectQuestionSchema = baseQuestionSchema.extend({
  type: z.literal('multi-select'), // Ensure the type is 'multi-select'
  minSelect: z.number().min(0, 'Minimum selection must be at least 0').optional(),
  maxSelect: z.number().min(1, 'Maximum selection must be at least 1').optional(),
  options: z
    .array(z.string().min(1, 'Option cannot be an empty string'))
    .min(1, 'Options cannot be empty'),
})

// Schema for file upload questions
export const fileUploadQuestionSchema = baseQuestionSchema.extend({
  type: z.literal('file-upload'), // Ensure the type is 'file-upload'
  allowedFileTypes: z.string().optional(),
  maxFileSizeMB: z.number().min(1, 'Maximum file size must be at least 1 MB').optional(),
})

// Combine schemas using discriminated union
export const questionConfigSchema = z
  .discriminatedUnion('type', [
    multiSelectQuestionSchema,
    textQuestionSchema,
    fileUploadQuestionSchema,
  ])
  .refine(
    (data) => {
      // If accessibleForOtherPhases = false, no validation needed.
      if (!data.accessibleForOtherPhases) return true

      // Otherwise, require accessKey to exist and to have no spaces:
      return (
        typeof data.accessKey === 'string' &&
        data.accessKey.trim().length > 0 && // optional: require it to be non-empty
        !/\s/.test(data.accessKey) // no spaces
      )
    },
    {
      message:
        'If "accessibleForOtherPhases" is checked, "accessKey" must be provided and cannot contain spaces.',
      path: ['accessKey'], // Attach the error to "accessKey"
    },
  )

// Export TypeScript types
export type QuestionConfigFormData = z.infer<typeof questionConfigSchema>
export type QuestionConfigFormDataMultiSelect = z.infer<typeof multiSelectQuestionSchema>
export type QuestionConfigFormDataText = z.infer<typeof textQuestionSchema>
export type QuestionConfigFormDataFileUpload = z.infer<typeof fileUploadQuestionSchema>
