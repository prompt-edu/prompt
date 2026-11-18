import * as z from 'zod'

const FORBIDDEN_CHARACTERS = /[-/%?#\p{Cc}]/u

export const courseIdentifierPart = (label: string) =>
  z
    .string()
    .min(1, `${label} is required`)
    .refine(
      (val) => !FORBIDDEN_CHARACTERS.test(val),
      `${label} cannot contain "-", "/", "%", "?", "#" or control characters`,
    )
