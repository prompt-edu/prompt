import type { CreateOrg, UpdateOrg } from '@core/interfaces/org'
import { describe, expect, it } from 'vitest'
import {
  MAX_CONTACT_EMAIL_LENGTH,
  MAX_NAME_LENGTH,
  MAX_WEBSITE_LENGTH,
  trimOrgDetails,
  validateCreateOrg,
  validateOrgDetails,
  validateSlug,
} from './validateOrgForm'

const details: UpdateOrg = {
  name: 'Applied Education Technologies',
  school: '',
  university: '',
  website: '',
  contactEmail: '',
}

describe('validateSlug', () => {
  it.each(['ab', 'aet', 'chair-1', 'a1-b2-c3', 'lab404', 'cgx', 'x-cgx', 'a'.repeat(50)])(
    'accepts %s',
    (slug) => {
      expect(validateSlug(slug)).toBeUndefined()
    },
  )

  it.each([
    ['', 'required'],
    ['a', 'between'],
    ['a'.repeat(51), 'between'],
    ['AET', 'lowercase'],
    ['ä-chair', 'lowercase'],
    ['-aet', 'lowercase'],
    ['aet-', 'lowercase'],
    ['a--b', 'lowercase'],
    ['a_b', 'lowercase'],
    ['a b', 'lowercase'],
    ['cg', '"cg"'],
    ['cg-aet', '"cg"'],
    ['aet-cg', '"cg"'],
    ['org-cg-admin', '"cg"'],
  ])('rejects %j', (slug, reason) => {
    expect(validateSlug(slug)).toContain(reason)
  })
})

describe('validateOrgDetails', () => {
  it('accepts a name alone, since every other field is optional', () => {
    expect(validateOrgDetails(details)).toEqual({})
  })

  it('accepts every optional field when it is valid', () => {
    expect(
      validateOrgDetails({
        ...details,
        school: 'School of Computation, Information and Technology',
        university: 'Technical University of Munich',
        website: 'https://aet.cit.tum.de#contact',
        contactEmail: 'chair@example.com',
      }),
    ).toEqual({})
  })

  it('requires a name', () => {
    expect(validateOrgDetails({ ...details, name: '' }).name).toBe('A name is required.')
  })

  it('counts code points like the server, not UTF-16 units', () => {
    expect(validateOrgDetails({ ...details, name: '😀'.repeat(MAX_NAME_LENGTH) })).toEqual({})
    expect(validateOrgDetails({ ...details, name: '😀'.repeat(MAX_NAME_LENGTH + 1) }).name).toBe(
      `The name must be at most ${MAX_NAME_LENGTH} characters.`,
    )
  })

  it.each(['example.com', 'ftp://example.com', 'https://', '/relative', 'mailto:a@b.de'])(
    'rejects the website %j',
    (website) => {
      expect(validateOrgDetails({ ...details, website }).website).toBe(
        'Enter an absolute http or https URL.',
      )
    },
  )

  it('rejects an overlong website', () => {
    const website = `https://example.com/${'a'.repeat(MAX_WEBSITE_LENGTH)}`
    expect(validateOrgDetails({ ...details, website }).website).toContain('at most')
  })

  it.each(['chair', 'chair@', '@example.com', 'Chair <chair@example.com>', 'a b@example.com'])(
    'rejects the contact email %j',
    (contactEmail) => {
      expect(validateOrgDetails({ ...details, contactEmail }).contactEmail).toBeDefined()
    },
  )

  it('rejects an overlong contact email', () => {
    const contactEmail = `${'a'.repeat(MAX_CONTACT_EMAIL_LENGTH)}@example.com`
    expect(validateOrgDetails({ ...details, contactEmail }).contactEmail).toContain('at most')
  })
})

describe('validateCreateOrg', () => {
  const values: CreateOrg = { ...details, slug: 'aet', parentOrgID: null }

  it('accepts a valid org', () => {
    expect(validateCreateOrg(values)).toEqual({})
  })

  it('reports the slug together with the details', () => {
    const errors = validateCreateOrg({ ...values, slug: 'cg', name: '' })
    expect(Object.keys(errors).sort()).toEqual(['name', 'slug'])
  })
})

describe('trimOrgDetails', () => {
  it('trims every field, so an empty optional field is sent as empty and cleared', () => {
    expect(
      trimOrgDetails({
        name: '  AET ',
        school: ' ',
        university: '\tTUM\n',
        website: ' https://aet.cit.tum.de ',
        contactEmail: '',
      }),
    ).toEqual({
      name: 'AET',
      school: '',
      university: 'TUM',
      website: 'https://aet.cit.tum.de',
      contactEmail: '',
    })
  })
})
