import { describe, expect, it } from 'vitest'
import { isOrgId } from './isOrgId'

describe('isOrgId', () => {
  it.each(['7a000001-0000-0000-0000-000000000001', '7A000001-0000-0000-0000-00000000ABCD'])(
    'accepts %s',
    (orgId) => {
      expect(isOrgId(orgId)).toBe(true)
    },
  )

  it.each([
    ['the own-orgs route', 'self'],
    ['a word', 'own'],
    ['an empty id', ''],
    ['no id', undefined],
    ['a UUID with a suffix', '7a000001-0000-0000-0000-000000000001x'],
    ['a UUID without hyphens', '7a000001000000000000000000000001'],
  ])('rejects %s', (_case, orgId) => {
    expect(isOrgId(orgId)).toBe(false)
  })
})
