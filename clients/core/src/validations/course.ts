import * as z from 'zod'

export const courseFormSchema = z.object({
  name: z
    .string()
    .min(1, 'Course name is required')
    .refine((val) => !/[-/]/.test(val), 'Course name cannot contain "-" or "/"'),
  dateRange: z.object({
    from: z.date(),
    to: z.date(),
  }),
  courseType: z.string().min(1, 'Course type is required'),
  ects: z.number().min(0, 'ECTS must be a positive number'),
  semesterTag: z
    .string()
    .min(1, 'Semester tag is required')
    .refine((val) => !/[-/]/.test(val), 'Semester tag cannot contain "-" or "/"'),
  shortDescription: z
    .string()
    .min(1, 'Short description is required')
    .max(255, 'Short description cannot exceed 255 characters'),
  longDescription: z
    .string()
    .max(5000, 'Long description cannot exceed 5000 characters')
    .optional()
    .or(z.literal('')),
})

export type CourseFormValues = z.infer<typeof courseFormSchema>
