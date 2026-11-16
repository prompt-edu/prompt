import { describe, expect, it } from 'vitest'
import { currentCursor, hasNewerPage, popNewerPage, pushOlderPage } from './paging'

describe('call paging', () => {
  it('starts on the newest page', () => {
    expect(currentCursor([])).toBeUndefined()
    expect(hasNewerPage([])).toBe(false)
  })

  it('walks to older pages and back', () => {
    const older = pushOlderPage([], 'a')
    const oldest = pushOlderPage(older, 'b')
    expect(currentCursor(oldest)).toBe('b')
    expect(hasNewerPage(oldest)).toBe(true)
    expect(popNewerPage(oldest)).toEqual(['a'])
  })

  it('stays on the last page when there is no older one', () => {
    expect(pushOlderPage(['a'], null)).toEqual(['a'])
  })
})
