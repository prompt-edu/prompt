type Id = string | undefined

export const infrastructureSetupKeys = {
  providerConfigs: (phaseId: Id) => ['provider-configs', phaseId] as const,
  providerAuthFields: (phaseId: Id, providerType: Id) =>
    ['provider-auth-fields', phaseId, providerType] as const,
  providerResourceTypes: (phaseId: Id, providerType: Id) =>
    ['provider-resource-types', phaseId, providerType] as const,
  resourceConfigs: (phaseId: Id) => ['resource-configs', phaseId] as const,
  setupConfig: (phaseId: Id) => ['setup-config', phaseId] as const,
  instances: (phaseId: Id) => ['instances', phaseId] as const,
  provisioningPreview: (phaseId: Id) => ['provisioning-preview', phaseId] as const,
  myResources: (phaseId: Id) => ['my-resources', phaseId] as const,
}
