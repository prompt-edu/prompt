import * as z from 'zod'

export const copyCourseSchema = z.object({
  name: z
    .string()
    .min(1, 'Course name is required')
    .refine((val) => !val.includes('-'), 'Course name cannot contain a "-" character'),
  dateRange: z
    .object({
      from: z.date(),
      to: z.date(),
    })
    .refine((data) => data.to > data.from, {
      message: 'End date must be after start date',
      path: ['to'],
    }),
  semesterTag: z
    .string()
    .min(1, 'Semester tag is required')
    .refine((val) => !val.includes('-'), 'Semester tag cannot contain a "-" character'),
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
