import { describe, expect, it } from 'vitest'
import { pageStarts } from './pageStarts'

describe('pageStarts', () => {
  it('keeps content that fits one page on a single page', () => {
    expect(pageStarts(800, 1000, [])).toEqual([0])
  })

  it('cuts at the page height when no block is in the way', () => {
    expect(pageStarts(2500, 1000, [])).toEqual([0, 1000, 2000])
  })

  it('moves the break up to the top of a block it would cut', () => {
    expect(pageStarts(1500, 1000, [{ top: 900, bottom: 1100 }])).toEqual([0, 900])
  })

  it('keeps the outermost fitting block together', () => {
    const blocks = [
      { top: 700, bottom: 1200 },
      { top: 950, bottom: 1050 },
    ]
    expect(pageStarts(1500, 1000, blocks)).toEqual([0, 700])
  })

  it('starts a new page at a forced break even when the content would fit', () => {
    expect(pageStarts(800, 1000, [], [500])).toEqual([0, 500])
  })

  it('places forced breaks before the regular page height', () => {
    expect(pageStarts(2500, 1000, [], [300, 1800])).toEqual([0, 300, 1300, 1800])
  })

  it('cuts through a block taller than a page', () => {
    expect(pageStarts(2500, 1000, [{ top: 100, bottom: 2400 }])).toEqual([0, 1000, 2000])
  })
})
