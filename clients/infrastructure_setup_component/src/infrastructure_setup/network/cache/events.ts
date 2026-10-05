import type { QueryClient } from '@tanstack/react-query'

import { infrastructureSetupKeys } from './keys'

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

// Every write changes what the next run would do, which the provisioning preview reports
export const infrastructureSetupCache = {
  providerSaved: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      infrastructureSetupKeys.providerConfigs(phaseId),
      infrastructureSetupKeys.provisioningPreview(phaseId),
    ]),

  // Resource configs cascade-delete with their provider, and instances with their config
  providerRemoved: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      infrastructureSetupKeys.providerConfigs(phaseId),
      infrastructureSetupKeys.resourceConfigs(phaseId),
      infrastructureSetupKeys.instances(phaseId),
      infrastructureSetupKeys.provisioningPreview(phaseId),
    ]),

  // Instances cascade-delete with their config, and each row carries its config's fields
  resourceConfigsChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      infrastructureSetupKeys.resourceConfigs(phaseId),
      infrastructureSetupKeys.instances(phaseId),
      infrastructureSetupKeys.provisioningPreview(phaseId),
    ]),

  setupConfigChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      infrastructureSetupKeys.setupConfig(phaseId),
      infrastructureSetupKeys.provisioningPreview(phaseId),
    ]),

  instancesChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      infrastructureSetupKeys.instances(phaseId),
      infrastructureSetupKeys.provisioningPreview(phaseId),
    ]),

  // A finished run turns its failures into retries
  provisioningRunEnded: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [infrastructureSetupKeys.provisioningPreview(phaseId)]),
}
