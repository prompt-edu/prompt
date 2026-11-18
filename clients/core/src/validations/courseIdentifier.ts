import * as z from 'zod'

const FORBIDDEN_CHARACTERS = /[-/%?#]/

export const courseIdentifierPart = (label: string) =>
  z
    .string()
    .min(1, `${label} is required`)
    .refine(
      (val) => !FORBIDDEN_CHARACTERS.test(val),
      `${label} cannot contain "-", "/", "%", "?" or "#"`,
    )
