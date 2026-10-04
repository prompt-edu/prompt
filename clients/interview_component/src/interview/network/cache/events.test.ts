import { QueryClient } from '@tanstack/react-query'
import { beforeEach, describe, expect, it } from 'vitest'

import { interviewCache } from './events'
import { interviewKeys } from './keys'

const PHASE = 'phase-1'
const OTHER_PHASE = 'phase-2'

let queryClient: QueryClient

const seed = (...keys: readonly (readonly unknown[])[]): void => {
  for (const key of keys) {
    queryClient.setQueryData(key, 'seeded')
  }
}

const isInvalidated = (key: readonly unknown[]): boolean =>
  queryClient.getQueryState(key)?.isInvalidated === true

beforeEach(() => {
  queryClient = new QueryClient()
})

describe('slotsChanged', () => {
  it('invalidates the slots of this phase only', () => {
    seed(interviewKeys.slots(PHASE), interviewKeys.slots(OTHER_PHASE))

    interviewCache.slotsChanged(queryClient, PHASE)

    expect(isInvalidated(interviewKeys.slots(PHASE))).toBe(true)
    expect(isInvalidated(interviewKeys.slots(OTHER_PHASE))).toBe(false)
  })

  it('leaves the reviews and the assignment alone', () => {
    seed(interviewKeys.reviews(PHASE), interviewKeys.myAssignment(PHASE))

    interviewCache.slotsChanged(queryClient, PHASE)

    expect(isInvalidated(interviewKeys.reviews(PHASE))).toBe(false)
    expect(isInvalidated(interviewKeys.myAssignment(PHASE))).toBe(false)
  })
})

describe('reviewWritten', () => {
  it('invalidates the reviews of this phase only', () => {
    seed(interviewKeys.reviews(PHASE), interviewKeys.reviews(OTHER_PHASE))

    interviewCache.reviewWritten(queryClient, PHASE)

    expect(isInvalidated(interviewKeys.reviews(PHASE))).toBe(true)
    expect(isInvalidated(interviewKeys.reviews(OTHER_PHASE))).toBe(false)
  })
})

describe('slotBooked', () => {
  it('invalidates my assignment and the slots whose counts it changes', () => {
    seed(interviewKeys.myAssignment(PHASE), interviewKeys.slots(PHASE))

    interviewCache.slotBooked(queryClient, PHASE)

    expect(isInvalidated(interviewKeys.myAssignment(PHASE))).toBe(true)
    expect(isInvalidated(interviewKeys.slots(PHASE))).toBe(true)
  })
})

describe('bookingCancelled', () => {
  it('clears my assignment at once, then fetches it and the slots again', async () => {
    let fetches = 0
    const fetchCounted = (queryKey: readonly unknown[]) =>
      queryClient.fetchQuery({
        queryKey,
        queryFn: () => {
          fetches += 1
          return 'fetched'
        },
      })
    await fetchCounted(interviewKeys.myAssignment(PHASE))
    await fetchCounted(interviewKeys.slots(PHASE))

    const cancelled = interviewCache.bookingCancelled(queryClient, PHASE)

    expect(queryClient.getQueryData(interviewKeys.myAssignment(PHASE))).toBeNull()
    await cancelled
    expect(fetches).toBe(4)
  })
})

describe('a key whose scoping id is missing', () => {
  it('is truncated at the missing segment rather than matching nothing', () => {
    seed(interviewKeys.slots(PHASE), interviewKeys.slots(OTHER_PHASE))

    interviewCache.slotsChanged(queryClient, undefined)

    expect(isInvalidated(interviewKeys.slots(PHASE))).toBe(true)
    expect(isInvalidated(interviewKeys.slots(OTHER_PHASE))).toBe(true)
  })

  it('keeps the namespace, so other caches stay untouched', () => {
    seed(interviewKeys.reviews(PHASE))

    interviewCache.slotsChanged(queryClient, undefined)

    expect(isInvalidated(interviewKeys.reviews(PHASE))).toBe(false)
  })
})
