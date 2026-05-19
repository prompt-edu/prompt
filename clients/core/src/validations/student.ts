import * as z from 'zod'
import { Gender, StudyDegree } from '@tumaet/prompt-shared-state'
import { translations } from '@tumaet/prompt-shared-state'

const universityLoginRegex = new RegExp(translations.university.universityLoginRegex)
const matriculationNumberRegex = new RegExp(translations.university.matriculationNumberRegex)

// Define the schema for a student form
export const studentBaseSchema = z.object({
  firstName: z.string().min(1, 'First name is required'),
  lastName: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email address'),
  gender: z.nativeEnum(Gender),
  nationality: z.string().min(1, 'Please select a nationality.'),
  studyProgram: z.string().min(1, 'Please select a study program.'),
  studyDegree: z.nativeEnum(StudyDegree),
  currentSemester: z
    .number()
    .int('Please enter your current semester')
    .min(1, 'Please select a semester.')
    .max(20, 'Please enter a number lower than or equal to 20'),
  hasUniversityAccount: z.literal(false), // Explicit literal for base case
})

// Define the schema for a university student form (extended)
export const studentUniversitySchema = z.object({
  firstName: z.string().min(1, 'First name is required'),
  lastName: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email address'),
  gender: z.nativeEnum(Gender),
  nationality: z.string().min(1, 'Please select a nationality.'),
  studyProgram: z.string().min(1, 'Please select a study program.'),
  studyDegree: z.nativeEnum(StudyDegree),
  currentSemester: z
    .number()
    .int('Please enter your current semester')
    .min(1, 'Please select a semester.')
    .max(20, 'Please enter a current semester number lower than or equal to 20'),
  hasUniversityAccount: z.literal(true), // Explicit literal for university case
  matriculationNumber: z
    .string()
    .refine(
      (val) => val === '' || matriculationNumberRegex.test(val),
      `Matriculation number must follow the pattern ${translations.university.matriculationExample}`,
    ),
  universityLogin: z
    .string()
    .regex(
      universityLoginRegex,
      `${translations.university.name} login should be of the format ${translations.university.universityLoginExample}`,
    ),
})

// Define the discriminated union based on `hasUniversityAccount`
export const studentSchema = z.discriminatedUnion('hasUniversityAccount', [
  studentBaseSchema,
  studentUniversitySchema,
])

export const questionConfigSchema = z.discriminatedUnion('hasUniversityAccount', [
  studentBaseSchema,
  studentUniversitySchema,
])

export type StudentFormValues = z.infer<typeof studentSchema>
export type StudentBaseFormValues = z.infer<typeof studentBaseSchema>
export type StudentUniversityFormValues = z.infer<typeof studentUniversitySchema>
