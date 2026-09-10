export type ProviderType = 'gitlab' | 'slack' | 'outline' | 'rancher' | 'keycloak'

export const providerTypes: ProviderType[] = ['gitlab', 'slack', 'outline', 'rancher', 'keycloak']

export interface ProviderConfig {
  id: string
  providerType: ProviderType
  // False for a provider copied from another phase, which keeps the row but not the
  // credentials. Such a provider cannot provision anything until they are re-entered.
  configured: boolean
}

export interface AuthField {
  name: string
  label: string
  type: 'text' | 'password'
  required: boolean
  description: string
}

export interface UpsertProviderConfigRequest {
  providerType: ProviderType
  credentials: Record<string, unknown>
}
