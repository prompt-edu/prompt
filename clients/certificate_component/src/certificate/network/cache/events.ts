import type { QueryClient } from '@tanstack/react-query'

import { certificateKeys } from './keys'

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

export const certificateCache = {
  configChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [certificateKeys.config(phaseId)]),

  // A download is recorded, and the list shows who has downloaded
  certificateDownloaded: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [certificateKeys.participants(phaseId)]),

  myCertificateDownloaded: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [certificateKeys.myStatus(phaseId)]),
}
