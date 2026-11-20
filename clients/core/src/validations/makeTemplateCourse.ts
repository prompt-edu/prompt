import * as z from 'zod'
import { courseIdentifierPart } from './courseIdentifier'

export const makeTemplateCourseSchema = z.object({
  name: courseIdentifierPart('Course name'),
  semesterTag: courseIdentifierPart('Semester tag'),
  shortDescription: z
    .string()
    .min(1, 'Short description is required')
    .max(255, 'Short description cannot exceed 255 characters'),
  longDescription: z
    .string()
    .max(5000, 'Long description cannot exceed 5000 characters')
    .optional(),
})

export type MakeTemplateCourseFormValues = z.infer<typeof makeTemplateCourseSchema>
