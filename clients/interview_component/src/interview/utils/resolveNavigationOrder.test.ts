import { describe, expect, it } from 'vitest'
import { resolveNavigationOrder } from './resolveNavigationOrder'

describe('resolveNavigationOrder', () => {
  const liveIds = ['a', 'b', 'c']

  it('keeps the snapshot order when it contains the current id', () => {
    expect(resolveNavigationOrder(['c', 'a', 'b'], liveIds, 'a')).toEqual(['c', 'a', 'b'])
  })

  it('drops snapshot ids that no longer exist', () => {
    expect(resolveNavigationOrder(['c', 'x', 'a'], liveIds, 'a')).toEqual(['c', 'a'])
  })

  it('falls back to the live order without a usable snapshot', () => {
    expect(resolveNavigationOrder(undefined, liveIds, 'a')).toEqual(liveIds)
    expect(resolveNavigationOrder(['c', 'b'], liveIds, 'a')).toEqual(liveIds)
  })
})
