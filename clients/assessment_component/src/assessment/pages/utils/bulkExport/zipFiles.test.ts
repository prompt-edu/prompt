import { describe, expect, it } from 'vitest'
import { personFileBaseName, uniqueFileNames } from './zipFiles'

describe('personFileBaseName', () => {
  it('joins last and first name and keeps letters with diacritics', () => {
    expect(personFileBaseName({ firstName: 'Zoë', lastName: 'Müller' })).toBe('Müller_Zoë')
  })

  it('replaces characters that are unsafe in file names', () => {
    expect(personFileBaseName({ firstName: 'Ann Marie', lastName: "O'Neil/../x" })).toBe(
      'O_Neil_x_Ann_Marie',
    )
  })
})

describe('uniqueFileNames', () => {
  it('numbers files that would otherwise share a name', () => {
    const files = [
      { baseName: 'Doe_Jane', extension: 'pdf' },
      { baseName: 'Doe_Jane', extension: 'pdf' },
      { baseName: 'Doe_Jane', extension: 'json' },
    ]
    expect(uniqueFileNames(files)).toEqual(['Doe_Jane.pdf', 'Doe_Jane_2.pdf', 'Doe_Jane.json'])
  })
})
