import { MAX_SLUG_LENGTH } from './validateOrgForm'

export const deriveOrgSlug = (name: string): string =>
  name
    .toLowerCase()
    .normalize('NFKD')
    .replace(/([aou])\u0308/g, '$1e')
    .replace(/ß/g, 'ss')
    .replace(/\p{M}/gu, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, MAX_SLUG_LENGTH)
    .replace(/-+$/, '')
