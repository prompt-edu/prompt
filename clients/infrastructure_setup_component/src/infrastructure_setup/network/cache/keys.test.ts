import { describe, expect, it } from 'vitest'

import { infrastructureSetupKeys } from './keys'

const PHASE = 'phase-1'

describe('infrastructureSetupKeys', () => {
  it('builds the configuration keys', () => {
    expect(infrastructureSetupKeys.providerConfigs(PHASE)).toEqual(['provider-configs', PHASE])
    expect(infrastructureSetupKeys.resourceConfigs(PHASE)).toEqual(['resource-configs', PHASE])
    expect(infrastructureSetupKeys.setupConfig(PHASE)).toEqual(['setup-config', PHASE])
  })

  it('scopes the provider catalog keys by provider type', () => {
    expect(infrastructureSetupKeys.providerAuthFields(PHASE, 'gitlab')).toEqual([
      'provider-auth-fields',
      PHASE,
      'gitlab',
    ])
    expect(infrastructureSetupKeys.providerResourceTypes(PHASE, 'gitlab')).toEqual([
      'provider-resource-types',
      PHASE,
      'gitlab',
    ])
  })

  it('keeps a missing id in the key rather than coercing it', () => {
    expect(infrastructureSetupKeys.providerAuthFields(PHASE, undefined)).toEqual([
      'provider-auth-fields',
      PHASE,
      undefined,
    ])
  })

  it('builds the provisioning keys', () => {
    expect(infrastructureSetupKeys.instances(PHASE)).toEqual(['instances', PHASE])
    expect(infrastructureSetupKeys.provisioningPreview(PHASE)).toEqual([
      'provisioning-preview',
      PHASE,
    ])
    expect(infrastructureSetupKeys.myResources(PHASE)).toEqual(['my-resources', PHASE])
  })
})
