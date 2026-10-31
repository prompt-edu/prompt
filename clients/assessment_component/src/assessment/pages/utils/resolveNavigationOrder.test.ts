import { describe, expect, it } from 'vitest'
import { resolveNavigationOrder } from './resolveNavigationOrder'

describe('resolveNavigationOrder', () => {
  const liveIds = ['a', 'b', 'c']

  it('keeps the table order when it contains the current id', () => {
    expect(resolveNavigationOrder(['c', 'a', 'b'], liveIds, 'a')).toEqual(['c', 'a', 'b'])
  })

  it('keeps a filtered table order', () => {
    expect(resolveNavigationOrder(['c', 'a'], liveIds, 'c')).toEqual(['c', 'a'])
  })

  it('drops ids that are no longer participants', () => {
    expect(resolveNavigationOrder(['c', 'x', 'a'], liveIds, 'a')).toEqual(['c', 'a'])
  })

  it('falls back to all participants without a usable table order', () => {
    expect(resolveNavigationOrder(undefined, liveIds, 'a')).toEqual(liveIds)
    expect(resolveNavigationOrder(['c', 'b'], liveIds, 'a')).toEqual(liveIds)
  })

  it('falls back to all participants without a current id', () => {
    expect(resolveNavigationOrder(['c', 'a'], liveIds, undefined)).toEqual(liveIds)
  })
})
