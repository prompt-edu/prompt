import { describe, expect, it } from 'vitest'
import { deriveOrgSlug } from './deriveOrgSlug'
import { MAX_SLUG_LENGTH, validateSlug } from './validateOrgForm'

describe('deriveOrgSlug', () => {
  it.each([
    ['Applied Education Technologies', 'applied-education-technologies'],
    ['  Chair of  Software--Engineering!  ', 'chair-of-software-engineering'],
    ['Lehrstuhl für Rechnerarchitektur', 'lehrstuhl-fuer-rechnerarchitektur'],
    ['Größe & Übung', 'groesse-uebung'],
    ['École Polytechnique', 'ecole-polytechnique'],
    ['TUM CIT 2026', 'tum-cit-2026'],
    ['', ''],
    ['!!!', ''],
  ])('derives %j as %j', (name, slug) => {
    expect(deriveOrgSlug(name)).toBe(slug)
  })

  it('cuts at the maximum length without leaving a trailing hyphen', () => {
    const slug = deriveOrgSlug(`${'a'.repeat(MAX_SLUG_LENGTH - 1)} chair`)
    expect(slug).toBe('a'.repeat(MAX_SLUG_LENGTH - 1))
    expect(validateSlug(slug)).toBeUndefined()
  })

  it('leaves a reserved "cg" segment for the validation to report', () => {
    expect(deriveOrgSlug('Chair for CG')).toBe('chair-for-cg')
    expect(validateSlug('chair-for-cg')).toContain('"cg"')
  })
})
