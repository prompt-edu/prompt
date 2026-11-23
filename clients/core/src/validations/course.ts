import * as z from 'zod'
import { courseIdentifierPart } from './courseIdentifier'

export const courseFormSchema = z.object({
  name: courseIdentifierPart('Course name'),
  dateRange: z.object({
    from: z.date(),
    to: z.date(),
  }),
  courseType: z.string().min(1, 'Course type is required'),
  ects: z.number().min(0, 'ECTS must be a positive number'),
  semesterTag: courseIdentifierPart('Semester tag'),
  shortDescription: z
    .string()
    .min(1, 'Short description is required')
    .max(255, 'Short description cannot exceed 255 characters'),
  longDescription: z
    .string()
    .max(5000, 'Long description cannot exceed 5000 characters')
    .optional()
    .or(z.literal('')),
  orgID: z.string().optional(),
})

export type CourseFormValues = z.infer<typeof courseFormSchema>
