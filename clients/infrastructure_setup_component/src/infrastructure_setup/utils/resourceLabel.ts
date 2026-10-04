import type { ProviderType } from '../interfaces/providerConfig'

const PROVIDER_NAMES: Record<ProviderType, string> = {
  gitlab: 'GitLab',
  slack: 'Slack',
  outline: 'Outline',
  rancher: 'Rancher',
  keycloak: 'Keycloak',
}

export const providerName = (providerType: ProviderType): string =>
  PROVIDER_NAMES[providerType] ?? providerType

// Names a kind of resource the way people say it: "GitLab group", "Slack channel".
export const resourceLabel = (providerType: ProviderType, resourceType: string): string =>
  `${providerName(providerType)} ${resourceType}`
