import type { QueryClient } from '@tanstack/react-query'

import { selfTeamAllocationKeys } from './keys'

type Id = string | undefined
type CacheKeys = readonly (readonly unknown[])[]

// A key holding `undefined` matches no cached entry, so invalidate from its last defined segment
const definedPrefixOf = (queryKey: readonly unknown[]): readonly unknown[] => {
  const missing = queryKey.indexOf(undefined)
  return missing === -1 ? queryKey : queryKey.slice(0, missing)
}

const invalidate = (queryClient: QueryClient, keys: CacheKeys): void => {
  for (const queryKey of keys) {
    queryClient.invalidateQueries({ queryKey: definedPrefixOf(queryKey) })
  }
}

export const selfTeamAllocationCache = {
  // Members and tutors are served as part of each team
  teamsChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [selfTeamAllocationKeys.teams(phaseId)]),

  // The config reports whether a timeframe is set
  timeframeChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      selfTeamAllocationKeys.timeframe(phaseId),
      selfTeamAllocationKeys.config(phaseId),
    ]),
}
