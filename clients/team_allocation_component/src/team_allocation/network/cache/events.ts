import type { QueryClient } from '@tanstack/react-query'

import { teamAllocationKeys } from './keys'

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

const teamKeys = (phaseId: Id): CacheKeys => [
  teamAllocationKeys.teams(phaseId),
  teamAllocationKeys.tease.teams(phaseId),
]

const surveyContentKeys = (phaseId: Id): CacheKeys => [
  teamAllocationKeys.survey.form(phaseId),
  teamAllocationKeys.survey.statistics(phaseId),
  teamAllocationKeys.tease.students(phaseId),
  teamAllocationKeys.config(phaseId),
]

export const teamAllocationCache = {
  // Deleting a team cascades to its allocations and the survey preferences for it
  teamsChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      ...teamKeys(phaseId),
      teamAllocationKeys.allocations(phaseId),
      ...surveyContentKeys(phaseId),
    ]),

  skillsChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      teamAllocationKeys.skills(phaseId),
      teamAllocationKeys.tease.skills(phaseId),
      ...surveyContentKeys(phaseId),
    ]),

  surveyTimeframeChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      teamAllocationKeys.survey.timeframe(phaseId),
      teamAllocationKeys.survey.form(phaseId),
      teamAllocationKeys.config(phaseId),
    ]),

  // Tutors are served as part of each team
  tutorsImported: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, teamKeys(phaseId)),

  mySurveyResponseSubmitted: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [teamAllocationKeys.survey.myResponse(phaseId)]),
}
