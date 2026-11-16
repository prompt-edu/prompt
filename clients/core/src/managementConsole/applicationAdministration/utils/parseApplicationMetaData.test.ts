import { describe, expect, it } from 'vitest'
import { parseApplicationMetaData } from './parseApplicationMetaData'

describe('parseApplicationMetaData', () => {
  it('enforces the student login when the phase has not stored it yet', () => {
    expect(parseApplicationMetaData({}).universityLoginAvailable).toBe(true)
    expect(parseApplicationMetaData(undefined).universityLoginAvailable).toBe(true)
  })

  it('keeps a stored student login setting', () => {
    expect(
      parseApplicationMetaData({ universityLoginAvailable: false }).universityLoginAvailable,
    ).toBe(false)
    expect(
      parseApplicationMetaData({ universityLoginAvailable: true }).universityLoginAvailable,
    ).toBe(true)
  })

  it('keeps the other settings off when they are not stored', () => {
    expect(parseApplicationMetaData({})).toEqual({
      applicationStartDate: undefined,
      applicationEndDate: undefined,
      externalStudentsAllowed: false,
      universityLoginAvailable: true,
      autoAccept: false,
      useCustomScores: false,
      applicationMode: 'apply',
      welcomeText: undefined,
    })
  })
})
