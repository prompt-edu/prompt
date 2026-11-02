import { describe, expect, it } from 'vitest'
import { getGeneralSettingsValues, hasGeneralSettingsChanges } from './generalSettingsValues'

const stored = getGeneralSettingsValues({
  applicationStartDate: new Date(2026, 9, 1, 9, 30),
  applicationEndDate: new Date(2026, 9, 31, 18, 0),
  externalStudentsAllowed: false,
  universityLoginAvailable: true,
  autoAccept: false,
  applicationMode: 'apply',
})

describe('getGeneralSettingsValues', () => {
  it('splits the stored dates into days and times', () => {
    expect(stored.startTime).toBe('09:30')
    expect(stored.endTime).toBe('18:00')
  })

  it('falls back to the defaults when nothing is stored', () => {
    expect(getGeneralSettingsValues({})).toEqual({
      startDate: undefined,
      endDate: undefined,
      startTime: '00:00',
      endTime: '23:59',
      externalStudentsAllowed: false,
      universityLoginAvailable: false,
      autoAccept: false,
      mode: 'apply',
    })
  })
})

describe('hasGeneralSettingsChanges', () => {
  it('reports no changes for the stored values', () => {
    expect(hasGeneralSettingsChanges({ ...stored }, stored)).toBe(false)
  })

  it('ignores the time of a date that was picked again for the same day', () => {
    expect(hasGeneralSettingsChanges({ ...stored, startDate: new Date(2026, 9, 1) }, stored)).toBe(
      false,
    )
  })

  it('reports a changed day, time, switch, or mode', () => {
    expect(hasGeneralSettingsChanges({ ...stored, endDate: new Date(2026, 10, 1) }, stored)).toBe(
      true,
    )
    expect(hasGeneralSettingsChanges({ ...stored, startTime: '10:00' }, stored)).toBe(true)
    expect(hasGeneralSettingsChanges({ ...stored, autoAccept: true }, stored)).toBe(true)
    expect(hasGeneralSettingsChanges({ ...stored, mode: 'import' }, stored)).toBe(true)
  })
})
