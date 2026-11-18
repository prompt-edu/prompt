import * as z from 'zod'
import { courseIdentifierPart } from './courseIdentifier'

export const copyCourseSchema = z.object({
  name: courseIdentifierPart('Course name'),
  dateRange: z
    .object({
      from: z.date(),
      to: z.date(),
    })
    .refine((data) => data.to > data.from, {
      message: 'End date must be after start date',
      path: ['to'],
    }),
  semesterTag: courseIdentifierPart('Semester tag'),
  courseType: z.string().min(1, 'Course type is required'),
  shortDescription: z
    .string()
    .min(1, 'Short description is required')
    .max(255, 'Short description cannot exceed 255 characters'),
  longDescription: z
    .string()
    .max(5000, 'Long description cannot exceed 5000 characters')
    .optional(),
})

export type CopyCourseFormValues = z.infer<typeof copyCourseSchema>
