import { QueryClient } from '@tanstack/react-query'
import { beforeEach, describe, expect, it } from 'vitest'

import { selfTeamAllocationCache } from './events'
import { selfTeamAllocationKeys } from './keys'

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

describe('teamsChanged', () => {
  it('invalidates the teams of this phase only', () => {
    seed(selfTeamAllocationKeys.teams(PHASE), selfTeamAllocationKeys.teams(OTHER_PHASE))

    selfTeamAllocationCache.teamsChanged(queryClient, PHASE)

    expect(isInvalidated(selfTeamAllocationKeys.teams(PHASE))).toBe(true)
    expect(isInvalidated(selfTeamAllocationKeys.teams(OTHER_PHASE))).toBe(false)
  })

  it('leaves the timeframe and the config alone', () => {
    seed(selfTeamAllocationKeys.timeframe(PHASE), selfTeamAllocationKeys.config(PHASE))

    selfTeamAllocationCache.teamsChanged(queryClient, PHASE)

    expect(isInvalidated(selfTeamAllocationKeys.timeframe(PHASE))).toBe(false)
    expect(isInvalidated(selfTeamAllocationKeys.config(PHASE))).toBe(false)
  })
})

describe('timeframeChanged', () => {
  it('invalidates the timeframe and the config reporting it', () => {
    seed(selfTeamAllocationKeys.timeframe(PHASE), selfTeamAllocationKeys.config(PHASE))

    selfTeamAllocationCache.timeframeChanged(queryClient, PHASE)

    expect(isInvalidated(selfTeamAllocationKeys.timeframe(PHASE))).toBe(true)
    expect(isInvalidated(selfTeamAllocationKeys.config(PHASE))).toBe(true)
  })

  it('leaves the teams alone', () => {
    seed(selfTeamAllocationKeys.teams(PHASE))

    selfTeamAllocationCache.timeframeChanged(queryClient, PHASE)

    expect(isInvalidated(selfTeamAllocationKeys.teams(PHASE))).toBe(false)
  })
})

describe('a key whose scoping id is missing', () => {
  it('is truncated at the missing segment rather than matching nothing', () => {
    seed(selfTeamAllocationKeys.teams(PHASE), selfTeamAllocationKeys.teams(OTHER_PHASE))

    selfTeamAllocationCache.teamsChanged(queryClient, undefined)

    expect(isInvalidated(selfTeamAllocationKeys.teams(PHASE))).toBe(true)
    expect(isInvalidated(selfTeamAllocationKeys.teams(OTHER_PHASE))).toBe(true)
  })

  it('keeps the namespace, so other caches stay untouched', () => {
    seed(selfTeamAllocationKeys.timeframe(PHASE))

    selfTeamAllocationCache.teamsChanged(queryClient, undefined)

    expect(isInvalidated(selfTeamAllocationKeys.timeframe(PHASE))).toBe(false)
  })
})
