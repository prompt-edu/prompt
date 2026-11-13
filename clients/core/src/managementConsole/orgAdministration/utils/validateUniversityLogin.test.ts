import { describe, expect, it } from 'vitest'
import { MAX_UNIVERSITY_LOGIN_LENGTH, validateUniversityLogin } from './validateUniversityLogin'

describe('validateUniversityLogin', () => {
  it.each(['ab12cde', 'ada.lovelace', 'ada@example.com', 'a'.repeat(MAX_UNIVERSITY_LOGIN_LENGTH)])(
    'accepts %s',
    (login) => {
      expect(validateUniversityLogin(login)).toBeUndefined()
    },
  )

  it.each([
    ['an empty login', ''],
    ['a login that is too long', 'a'.repeat(MAX_UNIVERSITY_LOGIN_LENGTH + 1)],
    ['a login with a space', 'ab 12cde'],
    ['a login with a tab', 'ab\t12cde'],
    ['a login with a colon', 'ab:12cde'],
    ['a login with a control character', 'ab\u000012cde'],
  ])('rejects %s', (_case, login) => {
    expect(validateUniversityLogin(login)).toBeDefined()
  })
})
