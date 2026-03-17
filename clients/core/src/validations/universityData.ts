import translations from '@/lib/translations.json'

import * as z from 'zod'

const universityLoginRegex = new RegExp(translations.university.universityLoginRegex)
const matriculationNumberRegex = new RegExp(translations.university.matriculationNumberRegex)

export const formSchemaUniversityData = z.object({
  email: z.string().email('Invalid email address'),
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

export type UniversityDataFormValues = z.infer<typeof formSchemaUniversityData>
