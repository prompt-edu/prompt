import { describe, expect, it } from 'vitest'
import { currentCursor, hasNewerPage, popNewerPage, pushOlderPage } from './auditLogPaging'
import type { AuditCursor } from './interfaces/auditLog'

const cursor = (id: string): AuditCursor => ({ createdAt: '2026-08-28T12:00:00Z', id })

describe('auditLogPaging', () => {
  it('treats an empty stack as the newest page', () => {
    expect(currentCursor([])).toBeUndefined()
    expect(hasNewerPage([])).toBe(false)
  })

  it('walks to older pages and back', () => {
    let stack = pushOlderPage([], cursor('a'))
    expect(currentCursor(stack)).toEqual(cursor('a'))
    expect(hasNewerPage(stack)).toBe(true)

    stack = pushOlderPage(stack, cursor('b'))
    expect(currentCursor(stack)).toEqual(cursor('b'))

    stack = popNewerPage(stack)
    expect(currentCursor(stack)).toEqual(cursor('a'))

    stack = popNewerPage(stack)
    expect(stack).toEqual([])
    expect(hasNewerPage(stack)).toBe(false)
  })

  it('stays put when there is no older page to go to', () => {
    const stack = pushOlderPage([cursor('a')], null)
    expect(stack).toEqual([cursor('a')])
  })
})
