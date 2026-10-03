import { describe, expect, it } from 'vitest'
import { DEFAULT_SORT, parseSortParam, serializeSortParam } from './sortOptions'

describe('parseSortParam', () => {
  it('reads a known ascending option', () => {
    expect(parseSortParam('lastName:asc')).toBe('lastName')
  })

  it('treats a missing order as ascending', () => {
    expect(parseSortParam('firstName')).toBe('firstName')
  })

  it('uses only the first sorting segment', () => {
    expect(parseSortParam('interviewScore:asc,lastName:asc')).toBe('interviewScore')
  })

  it('falls back to the default for missing, unknown or descending values', () => {
    expect(parseSortParam(null)).toBe(DEFAULT_SORT)
    expect(parseSortParam('')).toBe(DEFAULT_SORT)
    expect(parseSortParam('email:asc')).toBe(DEFAULT_SORT)
    expect(parseSortParam('lastName:desc')).toBe(DEFAULT_SORT)
  })
})

describe('serializeSortParam', () => {
  it('omits the default option', () => {
    expect(serializeSortParam(DEFAULT_SORT)).toBeNull()
  })

  it('round-trips a non-default option', () => {
    const serialized = serializeSortParam('acceptanceStatus')
    expect(serialized).toBe('acceptanceStatus:asc')
    expect(parseSortParam(serialized)).toBe('acceptanceStatus')
  })
})
