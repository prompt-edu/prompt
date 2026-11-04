import { toHttpUrl } from '@core/announcementBanner/utils/toHttpUrl'
import type { CreateOrg, UpdateOrg } from '@core/interfaces/org'

export type OrgFormErrors = Partial<Record<keyof CreateOrg, string>>

export const MIN_SLUG_LENGTH = 2
export const MAX_SLUG_LENGTH = 50
export const MAX_NAME_LENGTH = 255
export const MAX_SCHOOL_LENGTH = 255
export const MAX_UNIVERSITY_LENGTH = 255
export const MAX_WEBSITE_LENGTH = 2048
export const MAX_CONTACT_EMAIL_LENGTH = 254

const SLUG_PATTERN = /^[a-z0-9]+(-[a-z0-9]+)*$/
const EMAIL_PATTERN = /^[^\s@<>]+@[^\s@<>]+$/

const length = (value: string): number => [...value].length

export const validateSlug = (slug: string): string | undefined => {
  if (slug === '') {
    return 'A slug is required.'
  }
  if (!SLUG_PATTERN.test(slug)) {
    return 'Use lowercase letters and digits, separated by single hyphens.'
  }
  if (length(slug) < MIN_SLUG_LENGTH || length(slug) > MAX_SLUG_LENGTH) {
    return `The slug must be between ${MIN_SLUG_LENGTH} and ${MAX_SLUG_LENGTH} characters.`
  }
  if (slug.split('-').includes('cg')) {
    return 'The slug must not contain a "cg" segment, which is reserved for custom group roles.'
  }
  return undefined
}

export const validateOrgDetails = (values: UpdateOrg): OrgFormErrors => {
  const errors: OrgFormErrors = {}
  if (values.name === '') {
    errors.name = 'A name is required.'
  } else if (length(values.name) > MAX_NAME_LENGTH) {
    errors.name = `The name must be at most ${MAX_NAME_LENGTH} characters.`
  }
  if (length(values.school) > MAX_SCHOOL_LENGTH) {
    errors.school = `The school must be at most ${MAX_SCHOOL_LENGTH} characters.`
  }
  if (length(values.university) > MAX_UNIVERSITY_LENGTH) {
    errors.university = `The university must be at most ${MAX_UNIVERSITY_LENGTH} characters.`
  }
  if (length(values.website) > MAX_WEBSITE_LENGTH) {
    errors.website = `The website must be at most ${MAX_WEBSITE_LENGTH} characters.`
  } else if (values.website !== '' && !toHttpUrl(values.website)) {
    errors.website = 'Enter an absolute http or https URL.'
  }
  if (length(values.contactEmail) > MAX_CONTACT_EMAIL_LENGTH) {
    errors.contactEmail = `The contact email must be at most ${MAX_CONTACT_EMAIL_LENGTH} characters.`
  } else if (values.contactEmail !== '' && !EMAIL_PATTERN.test(values.contactEmail)) {
    errors.contactEmail = 'Enter a plain email address, such as chair@example.com.'
  }
  return errors
}

export const validateCreateOrg = (values: CreateOrg): OrgFormErrors => {
  const errors = validateOrgDetails(values)
  const slugError = validateSlug(values.slug)
  return slugError ? { ...errors, slug: slugError } : errors
}

export const trimOrgDetails = (values: UpdateOrg): UpdateOrg => ({
  name: values.name.trim(),
  school: values.school.trim(),
  university: values.university.trim(),
  website: values.website.trim(),
  contactEmail: values.contactEmail.trim(),
})
