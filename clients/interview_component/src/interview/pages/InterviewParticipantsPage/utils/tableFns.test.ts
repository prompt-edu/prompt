import { describe, expect, it } from 'vitest'
import {
  compareNullableNumbers,
  compareNullableStrings,
  getInterviewDayLabel,
  getInterviewDayOptions,
  isUrl,
  matchesNumericRange,
  matchesSelect,
  NOT_SCHEDULED,
} from './tableFns'

const MARCH_20_9AM = new Date(2026, 2, 20, 9, 0).getTime()
const MARCH_20_10AM = new Date(2026, 2, 20, 10, 0).getTime()
const MARCH_21_2PM = new Date(2026, 2, 21, 14, 0).getTime()

describe('compareNullableNumbers', () => {
  it('orders numbers ascending', () => {
    expect([3, 1, 2].sort(compareNullableNumbers)).toEqual([1, 2, 3])
  })

  it('puts missing values last', () => {
    expect([null, 2, null, 1].sort(compareNullableNumbers)).toEqual([1, 2, null, null])
  })
})

describe('compareNullableStrings', () => {
  it('orders strings alphabetically and puts missing values last', () => {
    expect(['MI 01', null, 'HS 1'].sort(compareNullableStrings)).toEqual(['HS 1', 'MI 01', null])
  })
})

describe('matchesNumericRange', () => {
  it('matches everything with a score when no bounds are set', () => {
    expect(matchesNumericRange(3, undefined)).toBe(true)
    expect(matchesNumericRange(3, {})).toBe(true)
  })

  it('excludes missing scores unless only missing scores are requested', () => {
    expect(matchesNumericRange(null, {})).toBe(false)
    expect(matchesNumericRange(null, { noScore: true })).toBe(true)
    expect(matchesNumericRange(3, { noScore: true })).toBe(false)
  })

  it('applies inclusive min and max bounds', () => {
    expect(matchesNumericRange(2, { min: '2', max: '4' })).toBe(true)
    expect(matchesNumericRange(4, { min: '2', max: '4' })).toBe(true)
    expect(matchesNumericRange(1, { min: '2' })).toBe(false)
    expect(matchesNumericRange(5, { max: '4' })).toBe(false)
  })

  it('matches nothing when a bound is not a number', () => {
    expect(matchesNumericRange(3, { min: 'abc' })).toBe(false)
  })
})

describe('matchesSelect', () => {
  it('matches everything when no option is selected', () => {
    expect(matchesSelect('March 20th, 2026', undefined)).toBe(true)
    expect(matchesSelect('March 20th, 2026', [])).toBe(true)
  })

  it('matches only the selected options', () => {
    expect(matchesSelect(NOT_SCHEDULED, [NOT_SCHEDULED])).toBe(true)
    expect(matchesSelect('March 20th, 2026', [NOT_SCHEDULED])).toBe(false)
  })
})

describe('getInterviewDayLabel', () => {
  it('formats the day of the interview', () => {
    expect(getInterviewDayLabel(MARCH_20_9AM)).toBe('March 20th, 2026')
  })

  it('labels a missing slot as not scheduled', () => {
    expect(getInterviewDayLabel(null)).toBe(NOT_SCHEDULED)
  })
})

describe('getInterviewDayOptions', () => {
  it('lists each interview day once, in chronological order', () => {
    expect(getInterviewDayOptions([MARCH_21_2PM, MARCH_20_10AM, MARCH_20_9AM])).toEqual([
      'March 20th, 2026',
      'March 21st, 2026',
    ])
  })

  it('appends the not scheduled option only when a participant has no slot', () => {
    expect(getInterviewDayOptions([MARCH_20_9AM, null])).toEqual([
      'March 20th, 2026',
      NOT_SCHEDULED,
    ])
  })
})

describe('isUrl', () => {
  it('recognizes http and https links', () => {
    expect(isUrl('https://tum-conf.zoom.us/j/123')).toBe(true)
    expect(isUrl('http://example.com')).toBe(true)
  })

  it('treats room names as plain text', () => {
    expect(isUrl('MI 00.08.038')).toBe(false)
  })
})
