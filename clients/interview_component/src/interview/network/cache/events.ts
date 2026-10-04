import type { QueryClient } from '@tanstack/react-query'

import { interviewKeys } from './keys'

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

export const interviewCache = {
  slotsChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [interviewKeys.slots(phaseId)]),

  reviewWritten: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [interviewKeys.reviews(phaseId)]),

  slotBooked: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [interviewKeys.myAssignment(phaseId), interviewKeys.slots(phaseId)]),

  bookingCancelled: async (queryClient: QueryClient, phaseId: Id): Promise<void> => {
    queryClient.setQueryData(interviewKeys.myAssignment(phaseId), null)
    await Promise.all(
      [interviewKeys.myAssignment(phaseId), interviewKeys.slots(phaseId)].map((queryKey) =>
        queryClient.refetchQueries({ queryKey }),
      ),
    )
  },
}
