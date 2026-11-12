import { MAX_SLUG_LENGTH } from './validateOrgForm'

// Letters NFKD leaves whole, which would otherwise become hyphens
const UNDECOMPOSABLE_LETTERS: Record<string, string> = {
  æ: 'ae',
  œ: 'oe',
  ø: 'o',
  ł: 'l',
  đ: 'd',
  ð: 'd',
  þ: 'th',
  ı: 'i',
}

const UNDECOMPOSABLE_LETTER_PATTERN = new RegExp(
  `[${Object.keys(UNDECOMPOSABLE_LETTERS).join('')}]`,
  'g',
)

export const deriveOrgSlug = (name: string): string =>
  name
    .toLowerCase()
    .normalize('NFKD')
    .replace(/([aou])\u0308/g, '$1e')
    .replace(/ß/g, 'ss')
    .replace(UNDECOMPOSABLE_LETTER_PATTERN, (letter) => UNDECOMPOSABLE_LETTERS[letter])
    .replace(/\p{M}/gu, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, MAX_SLUG_LENGTH)
    .replace(/-+$/, '')
