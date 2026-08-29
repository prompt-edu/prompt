import { describe, expect, it } from 'vitest'

import { computeQuartile } from './computeQuartile'

describe('computeQuartile', () => {
  it('returns 0 for an empty set', () => {
    expect(computeQuartile([], 0.5)).toBe(0)
  })

  it('returns the only value of a single-element set', () => {
    expect(computeQuartile([4], 0.25)).toBe(4)
  })

  it('sorts the values before picking a quartile', () => {
    const values = [9, 1, 7, 3, 5]

    expect(computeQuartile(values, 0.25)).toBe(3)
    expect(computeQuartile(values, 0.5)).toBe(5)
    expect(computeQuartile(values, 0.75)).toBe(7)
  })

  it('interpolates between the two neighbouring values', () => {
    expect(computeQuartile([1, 2, 3, 4], 0.5)).toBe(2.5)
    expect(computeQuartile([1, 2, 3, 4], 0.25)).toBe(1.75)
  })

  it('returns the bounds at the extremes', () => {
    expect(computeQuartile([1, 2, 3, 4], 0)).toBe(1)
    expect(computeQuartile([1, 2, 3, 4], 1)).toBe(4)
  })

  it('does not mutate the values it was given', () => {
    const values = [9, 1, 7]

    computeQuartile(values, 0.5)

    expect(values).toEqual([9, 1, 7])
  })
})
