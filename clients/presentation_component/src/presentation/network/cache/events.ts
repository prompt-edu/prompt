import type { QueryClient } from '@tanstack/react-query'

import { presentationKeys } from './keys'

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

// Materials and feedback cascade when a presentation is deleted
const presentationDataKeys = (phaseId: Id): CacheKeys => [
  presentationKeys.presentations.inPhase(phaseId),
  presentationKeys.materials.inPhase(phaseId),
  presentationKeys.feedback.inPhase(phaseId),
]

export const presentationCache = {
  // Unassigning a target deletes its presentation
  scheduleChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      presentationKeys.slots(phaseId),
      presentationKeys.targets(phaseId),
      ...presentationDataKeys(phaseId),
    ]),

  // A reset deletes the phase's presentations and slots, or its feedback forms
  settingsChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      presentationKeys.config(phaseId),
      presentationKeys.categories(phaseId),
      presentationKeys.slots(phaseId),
      presentationKeys.targets(phaseId),
      ...presentationDataKeys(phaseId),
    ]),

  materialsChanged: (queryClient: QueryClient, phaseId: Id, presentationId: Id): void =>
    invalidate(queryClient, [
      presentationKeys.materials.ofPresentation(phaseId, presentationId),
      presentationKeys.presentations.inPhase(phaseId),
    ]),

  feedbackChanged: (queryClient: QueryClient, phaseId: Id, presentationId: Id): void =>
    invalidate(queryClient, [presentationKeys.feedback.ofPresentation(phaseId, presentationId)]),

  feedbackStatusChanged: (queryClient: QueryClient, phaseId: Id, presentationId: Id): void =>
    invalidate(queryClient, [
      presentationKeys.feedback.ofPresentation(phaseId, presentationId),
      presentationKeys.presentations.inPhase(phaseId),
    ]),
}
