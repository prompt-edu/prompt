import { QueryClient } from '@tanstack/react-query'
import { beforeEach, describe, expect, it } from 'vitest'

import { infrastructureSetupCache } from './events'
import { infrastructureSetupKeys } from './keys'

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

const allPhaseKeys = (phaseId: string) => [
  infrastructureSetupKeys.providerConfigs(phaseId),
  infrastructureSetupKeys.providerAuthFields(phaseId, 'gitlab'),
  infrastructureSetupKeys.providerResourceTypes(phaseId, 'gitlab'),
  infrastructureSetupKeys.resourceConfigs(phaseId),
  infrastructureSetupKeys.setupConfig(phaseId),
  infrastructureSetupKeys.instances(phaseId),
  infrastructureSetupKeys.provisioningPreview(phaseId),
  infrastructureSetupKeys.myResources(phaseId),
]

const expectInvalidatedExactly = (expected: readonly (readonly unknown[])[]): void => {
  for (const key of allPhaseKeys(PHASE)) {
    const shouldBe = expected.some((e) => JSON.stringify(e) === JSON.stringify(key))
    expect(isInvalidated(key), JSON.stringify(key)).toBe(shouldBe)
  }
}

beforeEach(() => {
  queryClient = new QueryClient()
  seed(...allPhaseKeys(PHASE), ...allPhaseKeys(OTHER_PHASE))
})

describe('infrastructureSetupCache', () => {
  it('refreshes the providers and the preview when a provider is saved', () => {
    infrastructureSetupCache.providerSaved(queryClient, PHASE)

    expectInvalidatedExactly([
      infrastructureSetupKeys.providerConfigs(PHASE),
      infrastructureSetupKeys.provisioningPreview(PHASE),
    ])
  })

  it('refreshes everything that cascades when a provider is removed', () => {
    infrastructureSetupCache.providerRemoved(queryClient, PHASE)

    expectInvalidatedExactly([
      infrastructureSetupKeys.providerConfigs(PHASE),
      infrastructureSetupKeys.resourceConfigs(PHASE),
      infrastructureSetupKeys.instances(PHASE),
      infrastructureSetupKeys.provisioningPreview(PHASE),
    ])
  })

  it('refreshes the instances, which carry config fields, when a resource config changes', () => {
    infrastructureSetupCache.resourceConfigsChanged(queryClient, PHASE)

    expectInvalidatedExactly([
      infrastructureSetupKeys.resourceConfigs(PHASE),
      infrastructureSetupKeys.instances(PHASE),
      infrastructureSetupKeys.provisioningPreview(PHASE),
    ])
  })

  it('refreshes the setup config and the preview when the semester tag changes', () => {
    infrastructureSetupCache.setupConfigChanged(queryClient, PHASE)

    expectInvalidatedExactly([
      infrastructureSetupKeys.setupConfig(PHASE),
      infrastructureSetupKeys.provisioningPreview(PHASE),
    ])
  })

  it('refreshes the instances and the preview when an instance changes', () => {
    infrastructureSetupCache.instancesChanged(queryClient, PHASE)

    expectInvalidatedExactly([
      infrastructureSetupKeys.instances(PHASE),
      infrastructureSetupKeys.provisioningPreview(PHASE),
    ])
  })

  it('refreshes only the preview when a run ends', () => {
    infrastructureSetupCache.provisioningRunEnded(queryClient, PHASE)

    expectInvalidatedExactly([infrastructureSetupKeys.provisioningPreview(PHASE)])
  })

  it('leaves another phase alone', () => {
    infrastructureSetupCache.providerRemoved(queryClient, PHASE)

    for (const key of allPhaseKeys(OTHER_PHASE)) {
      expect(isInvalidated(key)).toBe(false)
    }
  })
})

describe('a key whose scoping id is missing', () => {
  it('is truncated at the missing segment rather than matching nothing', () => {
    infrastructureSetupCache.instancesChanged(queryClient, undefined)

    expect(isInvalidated(infrastructureSetupKeys.instances(PHASE))).toBe(true)
    expect(isInvalidated(infrastructureSetupKeys.instances(OTHER_PHASE))).toBe(true)
    expect(isInvalidated(infrastructureSetupKeys.resourceConfigs(PHASE))).toBe(false)
  })
})
